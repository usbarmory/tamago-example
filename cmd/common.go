// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"regexp"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"strconv"
	"time"

	"golang.org/x/term"
)

const testDiversifier = "\xde\xad\xbe\xef"

func init() {
	Add(Cmd{
		Name: "help",
		Help: "this help",
		Fn:   helpCmd,
	})

	Add(Cmd{
		Name:    "exit, quit",
		Args:    1,
		Pattern: regexp.MustCompile(`^(exit|quit)$`),
		Help:    "close session",
		Fn:      exitCmd,
	})

	Add(Cmd{
		Name: "stack",
		Help: "goroutine stack trace (current)",
		Fn:   stackCmd,
	})

	Add(Cmd{
		Name: "stackall",
		Help: "goroutine stack trace (all)",
		Fn:   stackallCmd,
	})

	Add(Cmd{
		Name:    "date",
		Args:    1,
		Pattern: regexp.MustCompile(`^date(.*)`),
		Syntax:  "(time in RFC339 format)?",
		Help:    "show/change runtime date and time",
		Fn:      dateCmd,
	})

	// The following commands are board specific, therefore their Fn
	// pointers are defined elsewhere in the respective target files.

	Add(Cmd{
		Name: "info",
		Help: "device information",
		Fn:   infoCmd,
	})

	Add(Cmd{
		Name: "reboot",
		Help: "reset device",
		Fn:   rebootCmd,
	})
}

func helpCmd(_ *Interface, term *term.Terminal, _ []string) (string, error) {
	return Help(term), nil
}

func exitCmd(_ *Interface, _ *term.Terminal, _ []string) (string, error) {
	log.Printf("Goodbye from %s/%s\n", runtime.GOOS, runtime.GOARCH)
	return "logout", io.EOF
}

func stackCmd(_ *Interface, _ *term.Terminal, _ []string) (string, error) {
	return string(debug.Stack()), nil
}

func stackallCmd(_ *Interface, _ *term.Terminal, _ []string) (string, error) {
	buf := new(bytes.Buffer)
	pprof.Lookup("goroutine").WriteTo(buf, 1)

	return buf.String(), nil
}

func dateCmd(_ *Interface, _ *term.Terminal, arg []string) (res string, err error) {
	if len(arg[0]) > 1 {
		t, err := time.Parse(time.RFC3339, arg[0][1:])

		if err != nil {
			return "", err
		}

		date(t.UnixNano())
	}

	return fmt.Sprintf("%s", time.Now().Format(time.RFC3339)), nil
}

func cipherCmd(arg []string, tag string, fn func(buf []byte) (string, error)) (res string, err error) {
	size, err := strconv.Atoi(arg[0])

	if err != nil {
		return "", fmt.Errorf("invalid size, %v", err)
	}

	sec, err := strconv.Atoi(arg[1])

	if err != nil {
		return "", fmt.Errorf("invalid duration, %v", err)
	}

	log.Printf("Doing %s for %ds on %d size blocks", tag, sec, size)

	n := 0
	buf := make([]byte, size)

	start := time.Now()
	duration := time.Duration(sec) * time.Second

	for time.Since(start) < duration {
		if _, err = fn(buf); err != nil {
			return
		}

		n++
	}

	elapsed := time.Since(start)
	kbps := (n * size) / int(elapsed/time.Millisecond)

	return fmt.Sprintf("%d %s's in %s (%dk)", n, tag, time.Since(start), kbps), nil
}
