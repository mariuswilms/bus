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

	_, msgs := b.Subscribe("athome1")
	b.Accept("athome1", "mydata")

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

	_, msgs := presence.Subscribe("athome")
	presence.Accept("athome", "mydata")
	time.Sleep(100 * time.Millisecond)

	select {
	case msg := <-msgs:
		t.Logf("Received message: %v", msg)
	default:
		t.Errorf("Expected to receive a message, but none was received")
	}
}
