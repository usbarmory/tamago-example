// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build sifive_u

package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"

	"github.com/usbarmory/tamago-example/shell"
	"github.com/usbarmory/tamago/board/qemu/sifive_u"
	"github.com/usbarmory/tamago/soc/sifive/fu540"
)

const boardName = "qemu-system-riscv64 (sifive_u)"

var NIC interface{}

func init() {
	Terminal = sifive_u.UART0
}

func date(epoch int64) {
	fu540.CLINT.SetTimer(epoch)
}

func uptime() (ns int64) {
	return fu540.CLINT.Nanotime() - fu540.CLINT.TimerOffset
}

func mem(start uint, size int, w []byte) (b []byte) {
	return memCopy(start, size, w)
}

func infoCmd(_ *shell.Interface, _ []string) (string, error) {
	var res bytes.Buffer

	ramStart, ramEnd := runtime.MemRegion()

	fmt.Fprintf(&res, "Runtime ......: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(&res, "RAM ..........: %#08x-%#08x (%d MiB)\n", ramStart, ramEnd, (ramEnd-ramStart)/(1024*1024))
	fmt.Fprintf(&res, "Board ........: %s\n", boardName)
	fmt.Fprintf(&res, "SoC ..........: %s\n", Target())

	return res.String(), nil
}

func rebootCmd(_ *shell.Interface, _ []string) (_ string, err error) {
	return "", errors.New("unimplemented")
}

func cryptoTest() {
	spawn(btcdTest)
	spawn(kemTest)
}

func storageTest() {
	return
}

func HasNetwork() (usb bool, eth bool) {
	return false, false
}

func Target() (t string) {
	return fmt.Sprintf("%s %v MHz", fu540.Model(), float32(fu540.Freq())/1000000)
}
