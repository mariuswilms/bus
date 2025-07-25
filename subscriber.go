// Copyright 2024 Marius Wilms. All rights reserved.
// Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
// Copyright 2018 Atelier Disko. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bus

import (
	"sync/atomic"
)

// subscriberID is a global counter for generating unique subscriber IDs.
var subscriberId atomic.Uint64

type Subscriber struct {
	receive chan<- Message
	topic   string
}

// Close will close the subscriber's channel. Safe to call multiple times.
//
// Usually the Broker will call this only once through unsubscribeAll or
// unsubscribe. In any case the subscriber will be removed from the broker,
// and subsequent calls cannot happen again, there is no subscriber to find
// anymore.
//
// However due to a race-condition between both methods, it is possible to
// call Close multiple times. This is safe and will not panic.
func (s *Subscriber) Close() error {
	if s.receive == nil {
		return nil
	}
	close(s.receive)
	s.receive = nil
	return nil
}
