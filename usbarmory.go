// https://github.com/f-secure-foundry/tamago-example
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// +build usbarmory

package main

import (
	"log"

	"github.com/f-secure-foundry/tamago/imx6"
	"github.com/f-secure-foundry/tamago/usbarmory/mark-two"
)

func init() {
	LED = usbarmory.LED
	SD = usbarmory.SD
	MMC = usbarmory.MMC

	if imx6.Native && (imx6.Family == imx6.IMX6UL || imx6.Family == imx6.IMX6ULL) {
		log.Println("-- i.mx6 ble ---------------------------------------------------------")
		usbarmory.BLE.Init()
		log.Println("ANNA-B112 BLE module initialized")
	}
}
