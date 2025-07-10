// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/usbarmory/tamago-example/shell"

	"github.com/usbarmory/tamago/board/qemu/microvm"
)

const (
	testDiversifier = "\xde\xad\xbe\xef"
	maxBufferSize   = 102400

	separator     = "-"
	separatorSize = 80
)

var (
	mux  sync.Mutex
	exit chan bool
	gr   int
)

func init() {
	shell.Add(shell.Cmd{
		Name: "test",
		Help: "launch tests",
		Fn:   testCmd,
	})
}

func msg(format string, args ...interface{}) {
	s := strings.Repeat(separator, 2) + " "
	s += fmt.Sprintf(" CPU %d • ", microvm.AMD64.LAPIC.ID())
	s += fmt.Sprintf(format, args...)
	s += " " + strings.Repeat(separator, separatorSize-len(s))

	log.Println(s)
}

func spawn(fn func() (tag, res string)) {
	gr += 1

	go func() {
		tag, res := fn()

		mux.Lock()
		defer mux.Unlock()

		if len(tag) > 0 {
			msg(tag)
		}

		if len(res) > 0 {
			log.Print(res)
		}

		exit <- true
	}()
}

func wait() {
	for i := 1; i <= gr; i++ {
		<-exit
	}

	gr = 0
}

func testCmd(_ *shell.Interface, _ []string) (_ string, _ error) {
	exit = make(chan bool)
	start := time.Now()

	spawn(timerTest)
	spawn(wakeTest)
	spawn(sleepTest)
	spawn(fsTest)
	spawn(rngTest)

	// spawns on its own
	cryptoTest()

	msg("launched %d test goroutines", gr)

	wait()
	msg("completed all test goroutines (%s)", time.Since(start))

	memTest()
	storageTest()
	msg("completed all tests (%s)", time.Since(start))

	return
}
