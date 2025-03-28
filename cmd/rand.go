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

	"github.com/usbarmory/tamago-example/shell"
)

const (
	rngRounds = 10
	rngSize   = 32
)

func init() {
	shell.Add(shell.Cmd{
		Name: "rand",
		Help: "gather 32 random bytes",
		Fn:   randCmd,
	})
}

func randCmd(_ *shell.Interface, _ []string) (string, error) {
	buf := make([]byte, 32)
	rand.Read(buf)
	return fmt.Sprintf("%x", buf), nil
}

func rngTest() (tag string, res string) {
	tag = "rng"

	b := &strings.Builder{}
	l := log.New(b, "", 0)
	l.SetPrefix(log.Prefix())

	n := rngSize
	buf := make([]byte, rngSize)

	l.Printf("%d reads of %d random bytes", rngRounds, n)
	start := time.Now()

	for i := 0; i < rngRounds; i++ {
		rand.Read(buf)
		l.Printf("  %x", buf)
	}

	l.Printf("done (%s)", time.Since(start))

	n = rngSize * rngRounds * 10
	buf = make([]byte, n)

	l.Printf("single read of %d random bytes", n)
	start = time.Now()

	rand.Read(buf)

	l.Printf("  %x", buf[0:rngSize])
	l.Printf("  ...")
	l.Printf("  %x", buf[n-rngSize:len(buf)])
	l.Printf("done (%s)", time.Since(start))

	return tag, b.String()
}
