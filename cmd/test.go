// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"log"
	"time"

	"golang.org/x/term"
)

func init() {
	Add(Cmd{
		Name: "test",
		Help: "launch tests",
		Fn:   testCmd,
	})
}

func (iface *Interface) spawn(fn func() (tag, res string)) {
	iface.gr += 1

	go func() {
		tag, res := fn()

		iface.Lock()
		defer iface.Unlock()

		if len(tag) > 0 {
			msg(tag)
		}

		if len(res) > 0 {
			log.Print(res)
		}

		iface.exit <- true
	}()
}

func (iface *Interface) wait() {
	for i := 1; i <= iface.gr; i++ {
		<-iface.exit
	}

	iface.gr = 0
}

func testCmd(iface *Interface, _ *term.Terminal, _ []string) (_ string, _ error) {
	iface.exit = make(chan bool)
	start := time.Now()

	iface.spawn(timerTest)
	iface.spawn(wakeTest)
	iface.spawn(sleepTest)
	iface.spawn(fsTest)
	iface.spawn(rngTest)

	// spawns on its own
	iface.cryptoTest()

	msg("launched %d test goroutines", iface.gr)

	iface.wait()

	msg("completed all test goroutines (%s)", time.Since(start))

	memTest()
	storageTest()

	return
}
