// Copyright 2024 Marius Wilms. All rights reserved.
// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bus

import (
	"context"
	"fmt"
	"regexp"
	"sync"
)

const TopicSeparator string = ":"
const MaxPendingMessages int = 10

func NewBroker(ctx context.Context) *Broker {
	b := &Broker{
		subscribed: make(map[uint64]*Subscriber),
		incoming:   make(chan *Message, MaxPendingMessages),
	}

	go func() {
		for {
			select {
			case msg := <-b.incoming:
				b.NotifyAll(msg)
			case <-ctx.Done():
				debug("Closing message broker (received quit)...")
				b.UnsubscribeAll()
				return
			}
		}
	}()

	return b
}

// Broker is the main event bus that services inside the DSK
// backend subscribe to.
type Broker struct {
	sync.RWMutex
	subscribed map[uint64]*Subscriber
	incoming   chan *Message
}

// Publish a message for fan-out. Will never block. When the buffer is
// full the message will be discarded and not delivered.
func (b *Broker) Publish(topic string, data interface{}) (bool, uint64) {
	msg := &Message{
		Id:    messageId.Add(1),
		Topic: topic,
		Data:  data,
	}
	return b.accept(msg)
}

func (b *Broker) accept(msg *Message) (ok bool, id uint64) {
	select {
	case b.incoming <- msg:
		debugf("Bus: accept %d '%s'", msg.Id, msg.Topic)
		ok = true
	default:
		debugf("Bus: buffer full, discarded %d", msg.Id)
		ok = false
	}
	return ok, msg.Id
}

func (b *Broker) NotifyAll(msg *Message) {
	b.RLock()
	defer b.RUnlock()

	for id, sub := range b.subscribed {
		matched, _ := regexp.MatchString(sub.topic, msg.Topic)
		if !matched {
			continue
		}
		debugf("Bus: notify %d", msg.Id)

		select {
		case sub.receive <- msg:
			// Subscriber received.
		default:
			debugf("Bus: buffer of subscriber %d full, not delivered", id)
		}
	}
}

func (b *Broker) Subscribe(topic string) (uint64, <-chan *Message) {
	debugf("Bus: subscribe '%s'", topic)

	b.Lock()
	defer b.Unlock()

	id := subscriberId.Add(1)
	ch := make(chan *Message, 10)

	b.subscribed[id] = &Subscriber{receive: ch, topic: topic}
	return id, ch
}

func (b *Broker) SubscribeFn(ctx context.Context, topic string, fn func(*Message)) {
	go func() {
		id, msgs := b.Subscribe(topic)

		for {
			select {
			case msg, ok := <-msgs:
				if !ok {
					debug("Stopping subscriber (channel closed)...")
					b.Unsubscribe(id)
					return
				}
				fn(msg)
			case <-ctx.Done():
				debug("Stopping subscriber (received quit)...")
				b.Unsubscribe(id)
				return
			}
		}
	}()
}

func (b *Broker) Unsubscribe(id uint64) {
	b.Lock()
	defer b.Unlock()

	if _, ok := b.subscribed[id]; ok {
		b.subscribed[id].Close()
		delete(b.subscribed, id)
	}
}

func (b *Broker) UnsubscribeAll() {
	b.Lock()
	defer b.Unlock()

	for id := range b.subscribed {
		b.subscribed[id].Close()
		delete(b.subscribed, id)
	}
}

// Connect will pass a subscribable messages through into this broker. The ID of the message
// will stay the same, but the topic will be changed using the provided namespace.
func (b *Broker) Connect(ctx context.Context, o *Broker, ns string) {
	debugf("Bus: connect onto '%s'", ns)

	o.SubscribeFn(ctx, `.*`, func(msg *Message) {
		debugf("Bus: forward %d => '%s'", msg.Id, ns)

		var topic string
		if ns != "" {
			topic = fmt.Sprintf("%s%s%s", ns, TopicSeparator, msg.Topic)
		} else {
			topic = msg.Topic
		}

		// Create a new message, as we cannot change the passed by
		// reference original message without changing the message for
		// other subscribers.
		fwd := &Message{
			Id:    msg.Id,
			Topic: topic,
			Data:  msg.Data,
		}
		b.accept(fwd)
	})
}
