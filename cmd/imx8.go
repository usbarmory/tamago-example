// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build imx8mpevk

package cmd

import (
	"bytes"
	"fmt"
	"runtime"
	_ "unsafe"

	"github.com/usbarmory/tamago-example/internal/semihosting"
	"github.com/usbarmory/tamago-example/shell"
)

const (
	romStart = 0x00000000
	romSize  = 0x3f000
)

func init() {
	runtime.Exit = func(_ int32) {
		semihosting.Exit()
	}
}

func date(epoch int64) {
	// TODO
	// imx8mp.ARM64.SetTime(epoch)
}

func uptime() (ns int64) {
	// TODO
	// return imx8mp.ARM64.GetTime() - imx8mp.ARM64.TimerOffset
	return 0
}

func mem(start uint, size int, w []byte) (b []byte) {
	// TODO: temporarily map page zero if required
	return memCopy(start, size, w)
}

func infoCmd(_ *shell.Interface, _ []string) (string, error) {
	var res bytes.Buffer

	ramStart, ramEnd := runtime.MemRegion()
	name, freq := Target()

	fmt.Fprintf(&res, "Runtime ......: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(&res, "RAM ..........: %#08x-%#08x (%d MiB)\n", ramStart, ramEnd, (ramEnd-ramStart)/(1024*1024))
	fmt.Fprintf(&res, "Board ........: %s\n", boardName)
	fmt.Fprintf(&res, "SoC ..........: %s\n", name)
	fmt.Fprintf(&res, "Frequency ....: %v MHz\n", float32(freq)/1e6)

	if NIC != nil {
		fmt.Fprintf(&res, "ENET%d ........: %s %d\n", NIC.Index, NIC.MAC, NIC.Stats)
	}

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

func Target() (name string, freq uint32) {
	// TODO
	return "(emulated)", 0
}
