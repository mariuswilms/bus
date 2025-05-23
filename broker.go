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
)

const TopicSeparator string = ":"
const MaxPendingMessages int = 10

func NewBroker(ctx context.Context) *Broker {
	b := &Broker{
		Subscribable: &Subscribable{},
		incoming:     make(chan *Message, MaxPendingMessages),
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
	*Subscribable

	// Incoming messages are sent here.
	incoming chan *Message
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

// Connect will pass a subscribable messages through into this broker. The ID of the message
// will stay the same, but the topic will be changed using the provided namespace.
func (b *Broker) Connect(ctx context.Context, o *Subscribable, ns string) {
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
