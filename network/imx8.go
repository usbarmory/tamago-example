// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build imx8mpevk

package network

import (
	"log"
	"runtime/goos"

	"github.com/usbarmory/tamago/arm64"
	"github.com/usbarmory/tamago/soc/nxp/enet"
	"github.com/usbarmory/tamago/soc/nxp/imx8mp"

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

func startInterruptHandler(eth *enet.ENET, iface *gnet.Interface) {
	var buf []byte

	imx8mp.GIC.Init()
	imx8mp.GIC.EnableInterrupt(arm64.TIMER_IRQ)

	if eth != nil {
		buf = make([]byte, gnet.EthernetMaximumSize + gnet.MTU)
		imx8mp.GIC.EnableInterrupt(eth.IRQ)
	}

	isr := func() {
		irq := imx8mp.GIC.GetInterrupt()

		switch {
		case irq == arm64.TIMER_IRQ:
			imx8mp.ARM64.SetAlarm(0)
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

		imx8mp.ARM64.SetAlarm(pollUntil)
		imx8mp.ARM64.WaitInterrupt()
	}

	arm64.ServiceInterrupts(isr)
}

func Init(console *shell.Interface, hasUSB bool, hasEth bool, nic **enet.ENET) {
	var eth *enet.ENET
	var iface *gnet.Interface

	if hasUSB {
		panic("unsupported")
	}

	if hasEth {
		eth = imx8mp.ENET1

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

	startInterruptHandler(eth, iface)
}
