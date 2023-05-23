// https://github.com/usbarmory/tamago-example
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/dma"
)

// Override usbarmory pkg ramSize and `mem` allocation, as having concurrent
// USB and Ethernet interfaces requires more than what the iRAM can handle.

const (
	dmaSize = 0xa00000 // 10MB
	dmaStart = 0xa0000000 - dmaSize
)

//go:linkname ramSize runtime.ramSize
var ramSize uint = 0x20000000 - dmaSize // 512MB - 10MB

func init() {
	dma.Init(dmaStart, dmaSize)
}
