// https://github.com/usbarmory/tamago-example
//
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
	"math"
	"math/big"
	mathrand "math/rand"
	"regexp"
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
		Args: 0,
		Pattern: regexp.MustCompile(`^rand`),
		Help: "gather 32 random bytes",
		Fn: randCmd,
	})
}

func randCmd(_ *term.Terminal, _ []string) (string, error) {
	buf := make([]byte, 32)
	rand.Read(buf)
	return fmt.Sprintf("%x", buf), nil
}

func rngTest() {
	msg("rng (%d runs)", rngRounds)

	for i := 0; i < rngRounds; i++ {
		rng := make([]byte, rngSize)
		rand.Read(rng)
		log.Printf("%x", rng)
	}

	msg("rng benchmark")

	start := time.Now()

	for i := 0; i < rngCount; i++ {
		rng := make([]byte, rngSize)
		rand.Read(rng)
	}

	log.Printf("retrieved %d random bytes in %s", rngSize*rngCount, time.Since(start))

	seed, _ := rand.Int(rand.Reader, big.NewInt(int64(math.MaxInt64)))
	mathrand.Seed(seed.Int64())
}
