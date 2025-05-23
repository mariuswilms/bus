// Copyright 2024 Marius Wilms. All rights reserved.
// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bus

import (
	"fmt"
	"sync/atomic"
)

const MaxCallDepth int = 10

// messageId is a global counter for generating unique message IDs.
var messageId atomic.Uint64

type Message struct {
	// An unique Id, the message can be identified by. Messages from
	// multiple brokers can be mixed together and still be identified.
	Id uint64

	// Topic name, used by subscribers when they want to listen to a
	// specific subset of messages going through the broker.
	Topic string

	// Data, arbitrary additional data the message can transport.
	Data any
}

func (msg *Message) String() string {
	return fmt.Sprintf("%s (%s)", msg.Id, msg.Topic)
}
