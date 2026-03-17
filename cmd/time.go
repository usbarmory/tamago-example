// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const sleep = 100 * time.Millisecond

func timerTest() (tag string, res string) {
	start := time.Now()
	t := time.NewTimer(sleep)

	for now := range t.C {
		tag = fmt.Sprintf("timer event %v (actual %v)", sleep, now.Sub(start))
		break
	}

	return
}

func sleepTest() (tag string, res string) {
	start := time.Now()
	time.Sleep(sleep)

	tag = fmt.Sprintf("timer sleep %s (actual %v)", sleep, time.Since(start))

	return
}

func wakeTest() (tag string, res string) {
	start := time.Now()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTRAP)
	defer signal.Stop(c)

	go func() {
		time.Sleep(sleep)
		signal.Relay(syscall.SIGTRAP)
	}()

	if <-c != syscall.SIGTRAP {
		panic("unexpected signal")
	}

	tag = fmt.Sprintf("got signal after %v", time.Since(start))

	return
}
