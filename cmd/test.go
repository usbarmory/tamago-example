// https://github.com/usbarmory/tamago-example
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"time"

	"golang.org/x/term"
)

const sleep = 100 * time.Millisecond

var gr int
var exit = make(chan bool)

func init() {
	Add(Cmd{
		Name: "test",
		Help: "launch tests",
		Fn: testCmd,
	})
}

func spawn(fn func()) {
	gr += 1

	go func() {
		fn()
		exit <- true
	}()
}

func wait() {
	for i := 1; i <= gr; i++ {
		<-exit
	}

	gr = 0
}

func timerTest() {
	msg("timer start (wake up in %v)", sleep)

	start := time.Now()
	t := time.NewTimer(sleep)

	for now := range t.C {
		msg("timer woke up after %v", now.Sub(start))
		break
	}
}

func sleepTest() {
	msg("sleeping (wake up in %s)", sleep)

	start := time.Now()
	time.Sleep(sleep)

	msg("slept %s (%v)", sleep, time.Since(start))
}

func testCmd(t *term.Terminal, _ []string) (_ string, _ error) {
	start := time.Now()

	spawn(timerTest)
	spawn(sleepTest)
	spawn(fsTest)
	spawn(rngTest)
	spawn(cryptoTest)

	msg("launched %d test goroutines", gr)

	wait()

	msg("completed all goroutines (%s)", time.Since(start))

	memTest()
	mmcTest()

	return
}
