// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build firecracker

package cmd

import (
	"github.com/usbarmory/tamago/board/firecracker/microvm"
	"github.com/usbarmory/tamago/kvm/clock"
	intel_uart "github.com/usbarmory/tamago/soc/intel/uart"
)

const boardName = "firecracker"

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
	Terminal = &UART{
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

func Target() (name string, freq uint32) {
	return microvm.AMD64.Name(), microvm.AMD64.Freq()
}

func HasNetwork() (usb bool, eth bool) {
	return false, true
}
