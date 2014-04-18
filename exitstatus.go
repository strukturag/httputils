// Copyright 2014 struktur AG. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httputils

import (
	"os"
	"sync"
)

var exitStatus = 0
var exitMutex sync.Mutex
var exitFuncs []func()
var exitInProgressMutex sync.Mutex

// Atexit saves f to be run when ExitWithStatus is called.
//
// This function is not guaranteed to be threadsafe.
func Atexit(f func()) {
	exitFuncs = append(exitFuncs, f)
}

// SetExitStatus sets a process exit status which will be used by
// ExitWithStatus.
//
// This function is threadsafe.
func SetExitStatus(status int) {
	exitMutex.Lock()
	if exitStatus < status {
		exitStatus = status
	}
	exitMutex.Unlock()
}

// ExitWithStatus calls os.Exit with the status provided to SetExitStatus,
// calling any callback functions set using Atexit.
//
// This function is threadsafe.
func ExitWithStatus() {
	exitInProgressMutex.Lock()
	for _, f := range exitFuncs {
		f()
	}
	os.Exit(exitStatus)
	// We never come here.
	exitInProgressMutex.Unlock()
}
