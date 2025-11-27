// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build mx6ullevk || usbarmory

package network

import (
	"log"
	"runtime"

	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/soc/nxp/enet"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
	"github.com/usbarmory/tamago/soc/nxp/usb"

	"github.com/usbarmory/tamago-example/shell"
)

func startInterruptHandler(usb *usb.USB, eth *enet.ENET) {
	imx6ul.GIC.Init(true, false)
	imx6ul.GIC.EnableInterrupt(imx6ul.TIMER_IRQ, true)

	if usb != nil {
		imx6ul.GIC.EnableInterrupt(usb.IRQ, true)
	}

	if eth != nil {
		imx6ul.GIC.EnableInterrupt(eth.IRQ, true)
	}

	isr := func() {
		irq := imx6ul.GIC.GetInterrupt(true)

		switch {
		case irq == imx6ul.TIMER_IRQ:
			imx6ul.ARM.SetAlarm(0)
		case usb != nil && irq == usb.IRQ:
			handleUSBInterrupt(usb)
		case eth != nil && irq == eth.IRQ:
			handleEthernetInterrupt(eth)
		default:
			log.Printf("internal error, unexpected IRQ %d", irq)
		}
	}

	// optimize CPU idle management as IRQs are enabled
	runtime.Idle = func(pollUntil int64) {
		if pollUntil == 0 {
			return
		}

		imx6ul.ARM.SetAlarm(pollUntil)
		imx6ul.ARM.WaitInterrupt()
	}

	arm.ServiceInterrupts(isr)
}

func Init(console *shell.Interface, hasUSB bool, hasEth bool, nic **enet.ENET) {
	var usb *usb.USB
	var eth *enet.ENET

	if hasUSB {
		usb = startUSB(console)
	}

	if hasEth {
		eth = imx6ul.ENET2

		if !imx6ul.Native {
			eth = imx6ul.ENET1
		}

		startEth(eth, console, true)
		*nic = eth
	}

	startInterruptHandler(usb, eth)
}
