// https://github.com/usbarmory/tamago-example
//
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
	"github.com/usbarmory/tamago/soc/imx6"
	"github.com/usbarmory/tamago/soc/imx6/imx6ul"
)

const (
	romStart = 0x00000000
	romSize  = 0x17000
)

func Remote() bool {
	return imx6.Native && (imx6.Family == imx6.IMX6UL || imx6.Family == imx6.IMX6ULL)
}

func Target() (t string) {
	t = fmt.Sprintf("%s %d MHz", imx6.Model(), imx6.ARMFreq()/1000000)

	if !imx6.Native {
		t += " (emulated)"
	}

	return
}

func date(epoch int64) {
	imx6.ARM.SetTimerOffset(epoch)
}

func mem(start uint32, size int, w []byte) (b []byte) {
	// temporarily map page zero if required
	if z := uint32(1 << 20); start < z {
		imx6.ARM.ConfigureMMU(0, z, (arm.TTE_AP_001<<10)|arm.TTE_SECTION)
		defer imx6.ARM.ConfigureMMU(0, z, 0)
	}

	return memCopy(start, size, w)
}

func infoCmd(_ *term.Terminal, _ []string) (string, error) {
	var res bytes.Buffer

	rom := mem(romStart, romSize, nil)

	res.WriteString(fmt.Sprintf("Runtime ......: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH))
	res.WriteString(fmt.Sprintf("Board ........: %s\n", boardName))
	res.WriteString(fmt.Sprintf("SoC ..........: %s\n", Target()))
	res.WriteString(fmt.Sprintf("SDP ..........: %v\n", imx6ul.SDP()))
	res.WriteString(fmt.Sprintf("Secure boot ..: %v\n", imx6.SNVS()))
	res.WriteString(fmt.Sprintf("Boot ROM hash : %x\n", sha256.Sum256(rom)))

	return res.String(), nil
}
