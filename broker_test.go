// Copyright 2025 Marius Wilms. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bus

import (
	"context"
	"testing"
	"time"
)

func TestBrokerPubSub(t *testing.T) {
	b := NewBroker(context.Background())

	msgs, _ := b.Subscribe("athome1")
	b.Publish("athome1", "mydata")

	time.Sleep(100 * time.Millisecond)

	select {
	case msg := <-msgs:
		t.Logf("Received message: %v", msg)
	default:
		t.Errorf("Expected to receive a message, but none was received")
	}
}

func TestBrokerEmbedding(t *testing.T) {
	type TestPresenceDetector struct {
		*Broker
	}

	b := NewBroker(context.Background())

	presence := &TestPresenceDetector{
		Broker: b,
	}

	msgs, _ := presence.Subscribe("athome")
	presence.Publish("athome", "mydata")
	time.Sleep(100 * time.Millisecond)

	select {
	case msg := <-msgs:
		t.Logf("Received message: %v", msg)
	default:
		t.Errorf("Expected to receive a message, but none was received")
	}
}

func TestDoubleCloseSubscriber(t *testing.T) {
	// Create a subscriber directly
	ch := make(chan Message, 10)
	sub := &Subscriber{receive: ch, topic: "test"}

	// Call Close multiple times - this should NOT panic with the fix
	sub.Close()
	sub.Close() // This should NOT panic with the atomic fix

	// If we get here, the fix is working
	t.Log("Fix is working - no panic occurred")
}
