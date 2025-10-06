// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build imx8mpevk

package cmd

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	_ "unsafe"

	"github.com/usbarmory/tamago/arm64"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/soc/nxp/imx8mp"
	"github.com/usbarmory/tamago/soc/nxp/snvs"

	"github.com/usbarmory/tamago-example/internal/semihosting"
	"github.com/usbarmory/tamago-example/shell"
)

const (
	// Override standard memory allocation as having concurrent USB and
	// Ethernet interfaces requires more than what the iRAM can handle.
	dmaSize  = 0xa00000 // 10MB
	dmaStart = 0x60000000 - dmaSize

	romStart = 0x00000000
	romSize  = 0x3f000
)

//go:linkname ramSize runtime.ramSize
var ramSize uint = 0x20000000 - dmaSize // 512MB - 10MB

func init() {
	dma.Init(dmaStart, dmaSize)

	runtime.Exit = func(_ int32) {
		semihosting.Exit()
	}
}

func date(epoch int64) {
	imx8mp.ARM64.SetTime(epoch)
}

func uptime() (ns int64) {
	return imx8mp.ARM64.GetTime() - imx8mp.ARM64.TimerOffset
}

func mem(start uint, size int, w []byte) (b []byte) {
	// temporarily map page zero if required
	if z := uint32(1 << 20); uint32(start) < z {
		imx8mp.ARM64.ConfigureMMU(0, z, 0, (arm64.TTE_AP_001<<10)|arm64.TTE_SECTION)
		defer imx8mp.ARM64.ConfigureMMU(0, z, 0, 0)
	}

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
	spawn(caamTest)
	spawn(dcpTest)

	return
}

func Target() (name string, freq uint32) {
	name = imx8mp.Model()
	name += " (emulated)"

	freq = imx8mp.ARMFreq()

	return
}
