// Copyright (c) The TamaGo Authors. All Rights Reserved.
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
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hako/durafmt"

	"github.com/usbarmory/tamago-example/shell"
	"github.com/usbarmory/tamago/dma"
)

var Terminal io.ReadWriter

var (
	once sync.Once
	idle func(pollUntil int64)
)

func init() {
	shell.Add(shell.Cmd{
		Name: "build",
		Help: "build information",
		Fn:   buildInfoCmd,
	})

	shell.Add(shell.Cmd{
		Name:    "exit, quit",
		Args:    1,
		Pattern: regexp.MustCompile(`^(exit|quit)$`),
		Help:    "close session",
		Fn:      exitCmd,
	})

	shell.Add(shell.Cmd{
		Name: "halt",
		Help: "halt the machine",
		Fn:   haltCmd,
	})

	shell.Add(shell.Cmd{
		Name: "stack",
		Help: "goroutine stack trace (current)",
		Fn:   stackCmd,
	})

	shell.Add(shell.Cmd{
		Name: "stackall",
		Help: "goroutine stack trace (all)",
		Fn:   stackallCmd,
	})

	shell.Add(shell.Cmd{
		Name:    "cpuidle",
		Args:    1,
		Pattern: regexp.MustCompile(`^cpuidle (on|off)$`),
		Help:    "CPU idle time management control",
		Syntax:  "(on|off)?",
		Fn:      cpuidleCmd,
	})

	shell.Add(shell.Cmd{
		Name:    "dma",
		Args:    1,
		Pattern: regexp.MustCompile(`^dma(?: (free|used))?$`),
		Help:    "show allocation of default DMA region",
		Syntax:  "(free|used)?",
		Fn:      dmaCmd,
	})

	shell.Add(shell.Cmd{
		Name:    "date",
		Args:    1,
		Pattern: regexp.MustCompile(`^date(.*)`),
		Syntax:  "(time in RFC339 format)?",
		Help:    "show/change runtime date and time",
		Fn:      dateCmd,
	})

	shell.Add(shell.Cmd{
		Name: "uptime",
		Help: "show how long the system has been running",
		Fn:   uptimeCmd,
	})

	// The following commands are board specific, therefore their Fn
	// pointers are defined elsewhere in the respective target files.

	shell.Add(shell.Cmd{
		Name: "info",
		Help: "device information",
		Fn:   infoCmd,
	})

	shell.Add(shell.Cmd{
		Name: "reboot",
		Help: "reset device",
		Fn:   rebootCmd,
	})
}

func buildInfoCmd(_ *shell.Interface, _ []string) (string, error) {
	var res bytes.Buffer

	if bi, ok := debug.ReadBuildInfo(); ok {
		res.WriteString(bi.String())
	}

	return res.String(), nil
}

func exitCmd(console *shell.Interface, _ []string) (string, error) {
	fmt.Fprintf(console.Output, "Goodbye from %s/%s\n", runtime.GOOS, runtime.GOARCH)
	return "logout", io.EOF
}

func haltCmd(console *shell.Interface, _ []string) (string, error) {
	fmt.Fprintf(console.Output, "Goodbye from %s/%s\n", runtime.GOOS, runtime.GOARCH)

	time.AfterFunc(
		100*time.Millisecond,
		func() { runtime.Exit(0) },
	)

	return "halted", io.EOF
}

func stackCmd(_ *shell.Interface, _ []string) (string, error) {
	return string(debug.Stack()), nil
}

func stackallCmd(_ *shell.Interface, _ []string) (string, error) {
	buf := new(bytes.Buffer)
	pprof.Lookup("goroutine").WriteTo(buf, 1)

	return buf.String(), nil
}

func cpuidleCmd(_ *shell.Interface, arg []string) (string, error) {
	once.Do(func() {
		idle = runtime.Idle
	})

	switch arg[0] {
	case "on":
		runtime.Idle = idle
	case "off":
		runtime.Idle = nil
	}

	return "", nil
}

func dmaCmd(_ *shell.Interface, arg []string) (string, error) {
	var res []string

	if dma.Default() == nil {
		return "no default DMA region is present", nil
	}

	dump := func(blocks map[uint]uint, tag string) string {
		var r []string
		var t uint

		for addr, n := range blocks {
			t += n
			r = append(r, fmt.Sprintf("%#08x-%#08x %10d", addr, addr+n, n))
		}

		sort.Strings(r)
		r = append(r, fmt.Sprintf("%21s %10d bytes %s", "", t, tag))

		return strings.Join(r, "\n")
	}

	if arg[0] == "" || arg[0] == "free" {
		if blocks := dma.Default().FreeBlocks(); len(blocks) > 0 {
			res = append(res, dump(blocks, "free"))
		}
	}

	if arg[0] == "" || arg[0] == "used" {
		if blocks := dma.Default().UsedBlocks(); len(blocks) > 0 {
			res = append(res, dump(blocks, "used"))
		}
	}

	return strings.Join(res, "\n"), nil
}

func dateCmd(_ *shell.Interface, arg []string) (res string, err error) {
	if len(arg[0]) > 1 {
		t, err := time.Parse(time.RFC3339, arg[0][1:])

		if err != nil {
			return "", err
		}

		date(t.UnixNano())
	}

	return fmt.Sprintf("%s", time.Now().Format(time.RFC3339)), nil
}

func uptimeCmd(_ *shell.Interface, _ []string) (string, error) {
	return fmt.Sprintf("%s", durafmt.Parse(time.Duration(uptime())*time.Nanosecond)), nil
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
