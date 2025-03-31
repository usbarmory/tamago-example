// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build cloud_hypervisor

package cmd

import (
	"log"

	"github.com/usbarmory/tamago/board/cloud_hypervisor/vm"
	"github.com/usbarmory/tamago/kvm/clock"
)

const boardName = "cloud_hypervisor"

func init() {
	Terminal = vm.UART0

	// set date and time at boot
	vm.AMD64.SetTimer(kvmclock.Now().UnixNano())

	log.SetPrefix("\r")
}

func date(epoch int64) {
	vm.AMD64.SetTimer(epoch)
}

func uptime() (ns int64) {
	return int64(float64(vm.AMD64.TimerFn()) * vm.AMD64.TimerMultiplier)
}

func Target() (name string, freq uint32) {
	return vm.AMD64.Name(), vm.AMD64.Freq()
}

func HasNetwork() (usb bool, eth bool) {
	return false, true
}
