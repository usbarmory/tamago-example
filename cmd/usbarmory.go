// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build usbarmory
// +build usbarmory

package cmd

import (
	"time"

	"golang.org/x/term"

	usbarmory "github.com/usbarmory/tamago/board/usbarmory/mk2"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

var boardName string

func init() {
	boardName = usbarmory.Model()
	console = usbarmory.UART2

	if !imx6ul.Native {
		return
	}

	I2C = append(I2C, usbarmory.I2C1)

	switch boardName {
	case "UA-MKII-β", "UA-MKII-γ":
		// On these models the standard serial console (UART2) is
		// exposed through the debug accessory, which needs to be
		// enabled.
		debugConsole, _ := usbarmory.DetectDebugAccessory(250 * time.Millisecond)
		<-debugConsole

		usbarmory.BLE.Init()

		MMC = append(MMC, usbarmory.SD)
	}

	MMC = append(MMC, usbarmory.MMC)
}

func rebootCmd(_ *Interface, _ *term.Terminal, _ []string) (_ string, _ error) {
	usbarmory.Reset()
	return
}

func HasNetwork() (usb bool, eth bool) {
	return imx6ul.Native, boardName == "UA-MKII-LAN"
}
