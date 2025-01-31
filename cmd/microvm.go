// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build microvm

package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"

	"golang.org/x/term"

	"github.com/usbarmory/tamago/board/qemu/microvm"
	"github.com/usbarmory/tamago/kvm/clock"
	"github.com/usbarmory/virtio-net"
)

var boardName = "microvm"
var NIC *vnet.Net

func init() {
	uart = microvm.UART0

	// set date and time at boot
	t, err := kvmclock.Now()

	if err != nil {
		return
	}

	microvm.AMD64.SetTimer(t.UnixNano())
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

	if NIC != nil {
		fmt.Fprintf(&res, "VirtIO Net%d ..: %s\n", NIC.Index, NIC.Config().MAC)
	}

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
