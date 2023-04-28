// https://github.com/usbarmory/tamago-example
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build sifive_u
// +build sifive_u

package network

import (
	"log"
	"os"

	"github.com/usbarmory/tamago/soc/nxp/enet"
	"github.com/usbarmory/tamago/soc/nxp/usb"
)

func StartInterruptHandler(usb *usb.USB, eth *enet.ENET) {
	log.Fatal("unsupported")
}

func StartEth(_ consoleHandler, _ *os.File) (_ *enet.ENET) {
	log.Fatal("unsupported")
	return
}

func StartUSB(_ consoleHandler, _ *os.File) (_ *usb.USB) {
	log.Fatal("unsupported")
	return
}
