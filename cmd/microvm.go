// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build microvm

package cmd

import (
	"github.com/usbarmory/tamago/board/qemu/microvm"
	"github.com/usbarmory/tamago/kvm/clock"
)

const boardName = "microvm"

func init() {
	uart = microvm.UART0

	// set date and time at boot
	microvm.AMD64.SetTimer(kvmclock.Now().UnixNano())
}

func date(epoch int64) {
	microvm.AMD64.SetTimer(epoch)
}

func uptime() (ns int64) {
	return int64(float64(microvm.AMD64.TimerFn()) * microvm.AMD64.TimerMultiplier)
}

func Target() (t string) {
	return microvm.AMD64.Name()
}
