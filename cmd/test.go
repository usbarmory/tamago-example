// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"log"
	"sync"
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
		Fn:   testCmd,
	})
}

var mux sync.Mutex

func spawn(fn func() (tag, res string)) {
	gr += 1

	go func() {
		tag, res := fn()
		exit <- true

		mux.Lock()
		defer mux.Unlock()

		if len(tag) > 0 {
			msg(tag)
		}

		if len(res) > 0 {
			log.Print(res)
		}
	}()
}

func wait() {
	for i := 1; i <= gr; i++ {
		<-exit
	}

	gr = 0
}

func timerTest() (tag string, res string) {
	start := time.Now()
	t := time.NewTimer(sleep)

	for now := range t.C {
		tag = fmt.Sprintf("timer expiration %v (actual %v)", sleep, now.Sub(start))
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

func testCmd(_ *Interface, _ *term.Terminal, _ []string) (_ string, _ error) {
	start := time.Now()

	spawn(timerTest)
	spawn(sleepTest)
	spawn(fsTest)
	spawn(rngTest)
	// spawns on its own
	cryptoTest()

	msg("launched %d test goroutines", gr)

	wait()

	msg("completed all goroutines (%s)", time.Since(start))

	memTest()
	usdhcTest()

	return
}
