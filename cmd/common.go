// https://github.com/usbarmory/tamago-example
//
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
	"time"

	"golang.org/x/term"
)

func init() {
	Add(Cmd{
		Name: "help",
		Args: 0,
		Pattern: regexp.MustCompile(`^help`),
		Help: "this help",
		Fn: helpCmd,
	})

	Add(Cmd{
		Name: "exit, quit",
		Args: 1,
		Pattern: regexp.MustCompile(`^(exit|quit)`),
		Help: "close session",
		Fn: exitCmd,
	})

	Add(Cmd{
		Name: "stack",
		Args: 0,
		Pattern: regexp.MustCompile(`^stack$`),
 		Help: "stack trace of current goroutine",
		Fn: stackCmd,
	})

	Add(Cmd{
		Name: "stackall",
		Args: 0,
		Pattern: regexp.MustCompile(`^stackall`),
		Help: "stack trace of all goroutines",
		Fn: stackallCmd,
	})

	Add(Cmd{
		Name: "date",
		Args: 1,
		Pattern: regexp.MustCompile(`^date(.*)`),
		Syntax: "(time in RFC339 format)?",
		Help: "show/change runtime date and time",
		Fn: dateCmd,
	})

	// The following commands are board specific, therefore their Fn
	// pointers are defined elsewhere in the respective target files.

	Add(Cmd{
		Name: "info",
		Args: 0,
		Pattern: regexp.MustCompile(`^info`),
		Help: "device information",
		Fn: infoCmd,
	})

	Add(Cmd{
		Name: "reboot",
		Help: "reset device",
		Fn: rebootCmd,
	})

}

func helpCmd(t *term.Terminal, _ []string) (string, error) {
	return string(t.Escape.Cyan) + Help() + string(t.Escape.Reset), nil
}

func exitCmd(_ *term.Terminal, _ []string) (string, error) {
	log.Printf("Goodbye from %s/%s\n", runtime.GOOS, runtime.GOARCH)
	return "logout", io.EOF
}

func stackCmd(_ *term.Terminal, _ []string) (string, error) {
	return string(debug.Stack()), nil
}

func stackallCmd(_ *term.Terminal, _ []string) (string, error) {
	buf := new(bytes.Buffer)
	pprof.Lookup("goroutine").WriteTo(buf, 1)

	return buf.String(), nil
}

func dateCmd(_ *term.Terminal, arg []string) (res string, err error) {
	if len(arg[0]) > 1 {
		t, err := time.Parse(time.RFC3339, arg[0][1:])

		if err != nil {
			return "", err
		}

		date(t.UnixNano())
	}

	return fmt.Sprintf("%s", time.Now().Format(time.RFC3339)), nil
}
