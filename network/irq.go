// https://github.com/usbarmory/tamago-example
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build mx6ullevk || usbarmory
// +build mx6ullevk usbarmory

package network

import (
	"log"

	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/soc/nxp/enet"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
	"github.com/usbarmory/tamago/soc/nxp/usb"
)

func StartInterruptHandler(usb *usb.USB, eth *enet.ENET) {
	imx6ul.GIC.Init(true, false)

	if usb != nil {
		imx6ul.GIC.EnableInterrupt(usb.IRQ, true)
	}

	if eth != nil {
		imx6ul.GIC.EnableInterrupt(eth.IRQ, true)
	}

	isr := func() {
		irq, end := imx6ul.GIC.GetInterrupt(true)

		if end != nil {
			close(end)
		}

		switch {
		case usb != nil && irq == usb.IRQ:
			handleUSBInterrupt(usb)
		case eth != nil && irq == eth.IRQ:
			handleEthernetInterrupt(eth)
		default:
			log.Printf("internal error, unexpected IRQ %d", irq)
		}
	}

	arm.ServiceInterrupts(isr)
}
