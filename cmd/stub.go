// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !amd64 && !mx6ullevk && !imx8mpevk && !usbarmory && !sifive_u

package cmd

import (
	"bytes"
	"fmt"
	"log"
	"runtime"

	"github.com/usbarmory/tamago-example/shell"
)

type CommonCPU interface {
	SetTime(int64)
	GetTime() int64
}

var (
	// CPU represents a common CPU interface
	CPU CommonCPU
	// Reboot represents the `reboot` command function.
	Reboot func()
)

func rebootCmd(_ *shell.Interface, _ []string) (_ string, err error) {
	if Reboot != nil {
		Reboot()
	}

	return
}

func date(epoch int64) {
	CPU.SetTime(epoch)
}

func uptime() (ns int64) {
	log.Printf("unimplemented")
	return 0
}

func infoCmd(_ *shell.Interface, _ []string) (string, error) {
	var res bytes.Buffer

	ramStart, ramEnd := runtime.MemRegion()

	fmt.Fprintf(&res, "Runtime ......: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(&res, "RAM ..........: %#08x-%#08x (%d MiB)\n", ramStart, ramEnd, (ramEnd-ramStart)/(1024*1024))

	return res.String(), nil
}

func cryptoTest() {
	spawn(btcdTest)
	spawn(kemTest)

	return
}

func storageTest() {
	return
}
