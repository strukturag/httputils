// Copyright 2014 struktur AG. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httputils

import (
	"fmt"
	"log"
)

// LogErrorf logs its arguments to the standard logger, and
// sets the exit status of ExitWithStatus to 1.
func LogErrorf(format string, args ...interface{}) {
	log.Printf(format, args...)
	SetExitStatus(1)
}

// LogFatalf logs its arguments to the standard logger and
// exits immediately.
func LogFatalf(format string, args ...interface{}) {
	LogErrorf(format, args...)
	ExitWithStatus()
}

// LogPrintFatalf logs its arguments to the standard logger as well as the
// standard outout and exits immediately.
func LogPrintFatalf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	LogFatalf(format, args...)
}
