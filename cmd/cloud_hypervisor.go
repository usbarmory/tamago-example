// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build cloud_hypervisor

package cmd

import (
	"bytes"
	"fmt"
	"log"

	"github.com/usbarmory/tamago/board/cloud_hypervisor/vm"
	"github.com/usbarmory/tamago/kvm/clock"
	"github.com/usbarmory/tamago/soc/intel/pci"

	"github.com/usbarmory/tamago-example/shell"
)

const boardName = "cloud_hypervisor"

func init() {
	Terminal = vm.UART0

	// set date and time at boot
	vm.AMD64.SetTimer(kvmclock.Now().UnixNano())

	log.SetPrefix("\r")

	shell.Add(shell.Cmd{
		Name: "lspci",
		Help: "list PCI devices",
		Fn:   lspciCmd,
	})
}

func date(epoch int64) {
	vm.AMD64.SetTimer(epoch)
}

func uptime() (ns int64) {
	return int64(float64(vm.AMD64.TimerFn()) * vm.AMD64.TimerMultiplier)
}

func lspciCmd(_ *shell.Interface, arg []string) (string, error) {
	var res bytes.Buffer

	fmt.Fprintf(&res, "Bus Vendor Device Bar0\n")

	for i := 0; i < 256; i++ {
		for _, d := range pci.Devices(i) {
			fmt.Fprintf(&res, "%03d %04x   %04x   %#016x\n", i, d.Vendor, d.Device, d.BaseAddress(0))
		}
	}

	return res.String(), nil
}

func Target() (name string, freq uint32) {
	return vm.AMD64.Name(), vm.AMD64.Freq()
}

func HasNetwork() (usb bool, eth bool) {
	return false, true
}
