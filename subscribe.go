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
	receive chan<- *Message
	topic   string
}

func (s *Subscriber) Close() error {
	close(s.receive)
	return nil
}
