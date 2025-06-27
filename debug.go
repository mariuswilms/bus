// Copyright 2024 Marius Wilms. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bus

import (
	"fmt"
	"log/slog"
	"os"
)

func isDebugMode() bool {
	return os.Getenv("BUS_DEBUG") == "y"
}

func debug(v ...any) {
	if isDebugMode() {
		slog.Debug(fmt.Sprintf("Bus: %s", fmt.Sprint(v...)))
	}
}

func debugf(format string, v ...any) {
	if isDebugMode() {
		slog.Debug(fmt.Sprintf("Bus: %s", fmt.Sprintf(format, v...)))
	}
}
