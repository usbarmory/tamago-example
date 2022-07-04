// https://github.com/usbarmory/tamago-example
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build mx6ullevk
// +build mx6ullevk

package cmd

import (
	"golang.org/x/term"

	"github.com/usbarmory/tamago/board/nxp/mx6ullevk"
	"github.com/usbarmory/tamago/soc/imx6"
)

const boardName = "MCIMX6ULL-EVK"

func init() {
	console = mx6ullevk.UART1

	if !(imx6.Native && (imx6.Family == imx6.IMX6UL || imx6.Family == imx6.IMX6ULL)) {
		return
	}

	mx6ullevk.I2C1.Init()
	I2C = append(I2C, mx6ullevk.I2C1)

	mx6ullevk.I2C2.Init()
	I2C = append(I2C, mx6ullevk.I2C2)

	MMC = append(MMC, mx6ullevk.SD1)
	MMC = append(MMC, mx6ullevk.SD2)
}

func rebootCmd(_ *term.Terminal, _ []string) (_ string, _ error) {
	mx6ullevk.Reset()
	return
}
