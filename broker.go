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

type UnsubscribeFn func()

func NewBroker(ctx context.Context) *Broker {
	b := &Broker{
		incoming: make(chan Message, MaxPendingMessages),
	}

	go func() {
		for {
			select {
			case msg := <-b.incoming:
				b.notifyAll(msg)
			case <-ctx.Done():
				debug("closing broker (received quit)...")
				b.unsubscribeAll()
				return
			}
		}
	}()

	return b
}

// Broker is the main event bus that services inside the DSK
// backend subscribe to.
type Broker struct {
	subscribed sync.Map
	incoming   chan Message
}

// Publish a message for fan-out. Will never block. When the buffer is
// full the message will be discarded and not delivered.
func (b *Broker) Publish(topic string, data interface{}) (bool, uint64) {
	msg := Message{
		Id:    messageId.Add(1),
		Topic: topic,
		Data:  data,
	}
	return b.accept(msg)
}

func (b *Broker) accept(msg Message) (ok bool, id uint64) {
	select {
	case b.incoming <- msg:
		debugf("accept %d '%s'", msg.Id, msg.Topic)
		ok = true
	default:
		debugf("buffer full, discarded %d", msg.Id)
		ok = false
	}
	return ok, msg.Id
}

func (b *Broker) notifyAll(msg Message) {
	b.subscribed.Range(func(key, value interface{}) bool {
		sub := value.(*Subscriber)
		matched, _ := regexp.MatchString(sub.topic, msg.Topic)
		if !matched {
			return true
		}
		debugf("notify %d", msg.Id)

		select {
		case sub.receive <- msg:
			// Subscriber received.
		default:
			debugf("buffer of subscriber %d full, not delivered", key.(uint64))
		}
		return true
	})
}

// Subscribe subscribes to the given topic, the func will return an unsubscribe function
// and an open channel where messages for the topic are received.
func (b *Broker) Subscribe(topic string) (<-chan Message, UnsubscribeFn) {
	debugf("subscribe '%s'", topic)

	id := subscriberId.Add(1)
	ch := make(chan Message, 10)

	b.subscribed.Store(id, &Subscriber{receive: ch, topic: topic})

	return ch, func() { b.unsubscribe(id) }
}

func (b *Broker) SubscribeFn(ctx context.Context, topic string, fn func(Message)) UnsubscribeFn {
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		msgs, unsubscribe := b.Subscribe(topic)

		for {
			select {
			case msg, ok := <-msgs:
				if !ok {
					debug("stopping subscriber (channel closed)...")
					unsubscribe()
					return
				}
				fn(msg)
			case <-ctx.Done():
				debug("stopping subscriber (received quit)...")
				unsubscribe()
				return
			}
		}
	}()

	return func() { cancel() }
}

func (b *Broker) unsubscribe(id uint64) {
	if sub, ok := b.subscribed.LoadAndDelete(id); ok {
		sub.(*Subscriber).Close()
	}
}

func (b *Broker) unsubscribeAll() {
	b.subscribed.Range(func(key, value interface{}) bool {
		value.(*Subscriber).Close()
		b.subscribed.Delete(key)
		return true
	})
}

// Connect will pass a subscribable messages through into this broker. The ID of the message
// will stay the same, but the topic will be changed using the provided namespace.
func (b *Broker) Connect(ctx context.Context, o *Broker, ns string) {
	debugf("connect onto '%s'", ns)

	o.SubscribeFn(ctx, `.*`, func(msg Message) {
		debugf("forward %d => '%s'", msg.Id, ns)

		var topic string
		if ns != "" {
			topic = fmt.Sprintf("%s%s%s", ns, TopicSeparator, msg.Topic)
		} else {
			topic = msg.Topic
		}

		// Create a new message with the modified topic
		fwd := Message{
			Id:    msg.Id,
			Topic: topic,
			Data:  msg.Data,
		}
		b.accept(fwd)
	})
}
