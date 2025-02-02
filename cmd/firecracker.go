// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build firecracker

package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"

	"golang.org/x/term"

	"github.com/usbarmory/tamago/board/firecracker/microvm"
	"github.com/usbarmory/tamago/kvm/clock"
	intel_uart "github.com/usbarmory/tamago/soc/intel/uart"
)

var boardName = "firecracker"
var NIC interface{}

type UART struct {
	uart *intel_uart.UART
}

func (hw *UART) Write(buf []byte) (n int, _ error) {
	return hw.uart.Write(buf)
}

func (hw *UART) Read(buf []byte) (n int, _ error) {
	n, _ = hw.uart.Read(buf)

	// We need to workaround the fact that firecracker sends \n on Enter
	// instead of \r which breaks term.ReadLine().
	for i, _ := range buf {
		if buf[i] == '\n' {
			buf[i] = '\r'
		}
	}

	return
}

func init() {
	// Workaround for buggy firecracker COM1
	uart = &UART{
		uart: microvm.UART0,
	}

	// set date and time at boot
	microvm.AMD64.SetTimer(kvmclock.Now().UnixNano())
}

func date(epoch int64) {
	microvm.AMD64.SetTimer(epoch)
}

func uptime() (ns int64) {
	return int64(float64(microvm.AMD64.TimerFn()) * microvm.AMD64.TimerMultiplier)
}

func mem(start uint, size int, w []byte) (b []byte) {
	return memCopy(start, size, w)
}

func infoCmd(_ *Interface, _ *term.Terminal, _ []string) (string, error) {
	var res bytes.Buffer

	ramStart, ramEnd := runtime.MemRegion()

	fmt.Fprintf(&res, "Runtime ......: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(&res, "RAM ..........: %#08x-%#08x (%d MiB)\n", ramStart, ramEnd, (ramEnd-ramStart)/(1024*1024))
	fmt.Fprintf(&res, "Board ........: %s\n", boardName)
	fmt.Fprintf(&res, "CPU ..........: %s\n", Target())

	return res.String(), nil
}

func rebootCmd(_ *Interface, _ *term.Terminal, _ []string) (_ string, err error) {
	return "", errors.New("unimplemented")
}

func (iface *Interface) cryptoTest() {
	iface.spawn(btcdTest)
	iface.spawn(kemTest)
}

func storageTest() {
	return
}

func HasNetwork() (usb bool, eth bool) {
	return false, true
}

func Target() (t string) {
	return microvm.AMD64.Name()
}
