// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build mx6ullevk || usbarmory

package network

import (
	"log"
	"runtime/goos"

	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/soc/nxp/enet"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
	"github.com/usbarmory/tamago/soc/nxp/usb"

	"github.com/usbarmory/go-net"
	"github.com/usbarmory/tamago-example/shell"
)

func handleEthernetInterrupt(eth *enet.ENET, iface *gnet.Interface, buf []byte) {
	for {
		if n, err := eth.Receive(buf); err != nil || n == 0 {
			return
		}

		iface.Stack.RecvInboundPacket(buf)
	}
}

func startInterruptHandler(usb *usb.USB, eth *enet.ENET, iface *gnet.Interface) {
	var buf []byte

	imx6ul.GIC.Init(true, false)
	imx6ul.GIC.EnableInterrupt(arm.TIMER_IRQ, true)

	if usb != nil {
		imx6ul.GIC.EnableInterrupt(usb.IRQ, true)
	}

	if eth != nil {
		buf = make([]byte, gnet.EthernetMaximumSize + gnet.MTU)
		imx6ul.GIC.EnableInterrupt(eth.IRQ, true)
	}

	isr := func() {
		irq := imx6ul.GIC.GetInterrupt(true)

		switch {
		case irq == arm.TIMER_IRQ:
			imx6ul.ARM.SetAlarm(0)
		case usb != nil && irq == usb.IRQ:
			handleUSBInterrupt(usb)
		case eth != nil && irq == eth.IRQ:
			handleEthernetInterrupt(eth, iface, buf)
		default:
			log.Printf("internal error, unexpected IRQ %d", irq)
		}
	}

	// optimize CPU idle management as IRQs are enabled
	goos.Idle = func(pollUntil int64) {
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
	var iface *gnet.Interface

	if hasUSB {
		usb = startUSB(console)
	}

	if hasEth {
		eth = imx6ul.ENET2

		if !imx6ul.Native {
			eth = imx6ul.ENET1
		}

		*nic = eth
		err := eth.Init()

		if err != nil {
			log.Printf("could not initialize network device, %v", err)
			return
		}

		if iface, err = initStack(console, eth); err != nil {
			log.Printf("could not start network stack, %v", err)
			return
		}

		eth.Start()
		eth.EnableInterrupt(enet.IRQ_RXF)
	}

	startInterruptHandler(usb, eth, iface)
}
