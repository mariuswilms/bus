// Copyright 2024 Marius Wilms. All rights reserved.
// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bus

import (
	"context"
	"regexp"
	"sync"
	"sync/atomic"
)

// Subscriber is a global counter for generating unique subscriber IDs.
var subscriberId atomic.Uint64

type Subscriber struct {
	receive chan<- *Message
	topic   string
}

func (s *Subscriber) Close() error {
	close(s.receive)
	return nil
}

// Subscribable is meant to be embedded by other structs, that
// are acting as a message/event bus.
type Subscribable struct {
	sync.RWMutex

	// A map of channels currently subscribed to changes. Once we
	// receive a message from the tree we fan it out to all. Once a
	// channel is detected to be closed, we remove it.
	subscribed map[uint64]*Subscriber
}

func (s *Subscribable) NotifyAll(msg *Message) {
	s.RLock()
	defer s.RUnlock()

	for id, sub := range s.subscribed {
		matched, _ := regexp.MatchString(sub.topic, msg.Topic)
		if !matched {
			continue
		}
		debugf("Bus: notify %s", msg.Id)

		select {
		case sub.receive <- msg:
			// Subscriber received.
		default:
			debugf("Bus: buffer of subscriber %s full, not delivered", id)
		}
	}
}

// Subscribe to a given topic. The topic name is interpreted as a
// regular expression.
func (s *Subscribable) Subscribe(topic string) (uint64, <-chan *Message) {
	debugf("Bus: subscribe '%s'", topic)

	s.Lock()
	defer s.Unlock()

	id := subscriberId.Add(1)
	ch := make(chan *Message, 10)

	if s.subscribed == nil {
		s.subscribed = make(map[uint64]*Subscriber, 0)
	}
	s.subscribed[id] = &Subscriber{receive: ch, topic: topic}
	return id, ch
}

// SubscribeFn registers a handler func that is invoked when receiving messages
// on the given topic.
func (s *Subscribable) SubscribeFn(ctx context.Context, topic string, fn func(*Message)) {
	go func() {
		id, msgs := s.Subscribe(topic)

		for {
			select {
			case msg, ok := <-msgs:
				if !ok {
					debug("Stopping subscriber (channel closed)...")
					s.Unsubscribe(id)
					return
				}
				fn(msg)
			case <-ctx.Done():
				debug("Stopping subscriber (received quit)...")
				s.Unsubscribe(id)
				return
			}
		}
	}()
}

// Unsubscribe a subscriber by its ID.
func (s *Subscribable) Unsubscribe(id uint64) {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.subscribed[id]; ok {
		s.subscribed[id].Close()
		delete(s.subscribed, id)
	}
}

// UnsubscribeAll unsubscribes all subscribers.
func (s *Subscribable) UnsubscribeAll() {
	s.Lock()
	defer s.Unlock()

	for id := range s.subscribed {
		s.subscribed[id].Close()
		delete(s.subscribed, id)
	}
}
