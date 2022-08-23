// https://github.com/usbarmory/tamago-example
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build usbarmory
// +build usbarmory

package cmd

import (
	"log"
	"time"

	"golang.org/x/term"

	usbarmory "github.com/usbarmory/tamago/board/usbarmory/mk2"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

var boardName = "USB armory Mk II"

func init() {
	boardName = usbarmory.Model()
	console = usbarmory.UART2

	if !imx6ul.Native {
		return
	}

	imx6ul.SetARMFreq(900)

	I2C = append(I2C, usbarmory.I2C1)

	MMC = append(MMC, usbarmory.SD)
	MMC = append(MMC, usbarmory.MMC)

	// On the USB armory Mk II the standard serial console (UART2) is
	// exposed through the debug accessory, which needs to be enabled.
	debugConsole, _ := usbarmory.DetectDebugAccessory(250 * time.Millisecond)
	<-debugConsole

	usbarmory.BLE.Init()
	log.Println("ANNA-B112 BLE module initialized")
}

func rebootCmd(_ *term.Terminal, _ []string) (_ string, _ error) {
	usbarmory.Reset()
	return
}
