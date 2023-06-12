// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build mx6ullevk || usbarmory
// +build mx6ullevk usbarmory

package cmd

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"runtime"

	"golang.org/x/term"

	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

const (
	romStart = 0x00000000
	romSize  = 0x17000
)

func Target() (t string) {
	t = fmt.Sprintf("%s %v MHz", imx6ul.Model(), float32(imx6ul.ARMFreq())/1000000)

	if !imx6ul.Native {
		t += " (emulated)"
	}

	return
}

func init() {
	if !imx6ul.Native {
		return
	}

	switch imx6ul.Model() {
	case "i.MX6ULL", "i.MX6ULZ":
		imx6ul.SetARMFreq(imx6ul.FreqMax)
	case "i.MX6UL":
		imx6ul.SetARMFreq(imx6ul.Freq528)
	}
}

func date(epoch int64) {
	imx6ul.ARM.SetTimer(epoch)
}

func mem(start uint, size int, w []byte) (b []byte) {
	// temporarily map page zero if required
	if z := uint32(1 << 20); uint32(start) < z {
		imx6ul.ARM.ConfigureMMU(0, z, (arm.TTE_AP_001<<10)|arm.TTE_SECTION)
		defer imx6ul.ARM.ConfigureMMU(0, z, 0)
	}

	return memCopy(start, size, w)
}

func infoCmd(_ *Interface, _ *term.Terminal, _ []string) (string, error) {
	var res bytes.Buffer

	ramStart, ramEnd := runtime.MemRegion()
	rom := mem(romStart, romSize, nil)

	res.WriteString(fmt.Sprintf("Runtime ......: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH))
	res.WriteString(fmt.Sprintf("RAM ..........: %#08x-%#08x (%d MiB)\n", ramStart, ramEnd, (ramEnd-ramStart)/(1024*1024)))
	res.WriteString(fmt.Sprintf("Board ........: %s\n", boardName))
	res.WriteString(fmt.Sprintf("SoC ..........: %s\n", Target()))
	res.WriteString(fmt.Sprintf("SDP ..........: %v\n", imx6ul.SDP))
	res.WriteString(fmt.Sprintf("Secure boot ..: %v\n", imx6ul.HAB()))
	res.WriteString(fmt.Sprintf("Boot ROM hash : %x\n", sha256.Sum256(rom)))

	if imx6ul.Native {
		res.WriteString(fmt.Sprintf("Unique ID ....: %X\n", imx6ul.UniqueID()))
		res.WriteString(fmt.Sprintf("Temperature ..: %f\n", imx6ul.TEMPMON.Read()))
	}

	return res.String(), nil
}

func cryptoTest() {
	spawn(ecdsaTest)
	spawn(btcdTest)
	spawn(kyberTest)
	spawn(caamTest)
	spawn(dcpTest)
}
