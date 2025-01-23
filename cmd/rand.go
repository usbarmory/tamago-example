// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"crypto/rand"
	"fmt"
	"log"
	"strings"
	"time"

	"golang.org/x/term"
)

const (
	rngRounds = 10
	rngSize   = 32
)

func init() {
	Add(Cmd{
		Name: "rand",
		Help: "gather 32 random bytes",
		Fn:   randCmd,
	})
}

func randCmd(_ *Interface, _ *term.Terminal, _ []string) (string, error) {
	buf := make([]byte, 32)
	rand.Read(buf)
	return fmt.Sprintf("%x", buf), nil
}

func rngTest() (tag string, res string) {
	tag = "rng"

	b := &strings.Builder{}
	log := log.New(b, "", 0)

	n := rngSize
	buf := make([]byte, rngSize)

	log.Printf("%d reads of %d random bytes", rngRounds, n)
	start := time.Now()

	for i := 0; i < rngRounds; i++ {
		rand.Read(buf)
		log.Printf("  %x", buf)
	}

	log.Printf("done (%s)", time.Since(start))

	n = rngSize * rngRounds * 10
	buf = make([]byte, n)

	log.Printf("single read of %d random bytes", n)
	start = time.Now()

	rand.Read(buf)

	log.Printf("  %x\n  ...\n  %x", buf[0:rngSize], buf[n-rngSize:len(buf)])
	log.Printf("done (%s)", time.Since(start))

	return tag, b.String()
}
