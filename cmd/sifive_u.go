// https://github.com/usbarmory/tamago-example
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build sifive_u
// +build sifive_u

package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"

	"golang.org/x/term"

	"github.com/usbarmory/tamago/board/qemu/sifive_u"
	"github.com/usbarmory/tamago/soc/sifive/fu540"
)

var boardName = "qemu-system-riscv64 (sifive_u)"

func init() {
	console = sifive_u.UART1
}

func Remote() bool {
	return false
}

func Target() (t string) {
	return fmt.Sprintf("%s %v MHz", fu540.Model(), float32(fu540.Freq())/1000000)
}

func date(epoch int64) {
	fu540.CLINT.SetTimer(epoch)
}

func mem(start uint32, size int, w []byte) (b []byte) {
	return memCopy(start, size, w)
}

func infoCmd(_ *term.Terminal, _ []string) (string, error) {
	var res bytes.Buffer

	res.WriteString(fmt.Sprintf("Runtime ......: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH))
	res.WriteString(fmt.Sprintf("Board ........: %s\n", boardName))
	res.WriteString(fmt.Sprintf("SoC ..........: %s\n", Target()))

	return res.String(), nil
}

func rebootCmd(_ *term.Terminal, _ []string) (_ string, err error) {
	return "", errors.New("unimplemented")
}

func cryptoTest() {
	spawn(ecdsaTest)
	spawn(btcdTest)
}

func mmcTest() {
	return
}
