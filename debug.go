// Copyright 2024 Marius Wilms. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bus

import (
	"log"
	"os"
)

func isDebugMode() bool {
	return os.Getenv("BUS_DEBUG") == "y"
}

func debug(v ...interface{}) {
	if isDebugMode() {
		log.Print(v...)
	}
}

func debugf(format string, v ...interface{}) {
	if isDebugMode() {
		log.Printf(format, v...)
	}
}
