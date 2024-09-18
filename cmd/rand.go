// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
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
	rngCount  = 1000
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
	tag = fmt.Sprintf("rng (%d runs)", rngRounds)

	b := &strings.Builder{}
	log := log.New(b, "", 0)

	for i := 0; i < rngRounds; i++ {
		rng := make([]byte, rngSize)
		rand.Read(rng)
		log.Printf("%x", rng)
	}

	start := time.Now()

	for i := 0; i < rngCount; i++ {
		rng := make([]byte, rngSize)
		rand.Read(rng)
	}

	log.Printf("retrieved %d random bytes in %s", rngSize*rngCount, time.Since(start))

	return tag, b.String()
}
