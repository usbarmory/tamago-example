// https://github.com/f-secure-foundry/tamago-example
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build mx6ullevk || usbarmory
// +build mx6ullevk usbarmory

package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"runtime"

	"github.com/f-secure-foundry/tamago/soc/imx6"
	"github.com/f-secure-foundry/tamago/soc/imx6/usb"
)

const (
	romStart = 0x00000000
	romSize  = 0x17000
)

var i2c []*imx6.I2C

func info() string {
	var res bytes.Buffer

	rom := mem(romStart, romSize, nil)

	res.WriteString(fmt.Sprintf("Runtime ......: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH))
	res.WriteString(fmt.Sprintf("Board ........: %s\n", boardName))
	res.WriteString(fmt.Sprintf("SoC ..........: %s %d MHz\n", imx6.Model(), imx6.ARMFreq()/1000000))
	res.WriteString(fmt.Sprintf("SDP ..........: %v\n", usb.SDP()))
	res.WriteString(fmt.Sprintf("Secure boot ..: %v\n", imx6.SNVS()))
	res.WriteString(fmt.Sprintf("Boot ROM hash : %x\n", sha256.Sum256(rom)))

	return res.String()
}
