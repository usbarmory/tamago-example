// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build imx8mpevk

package network

import (
	"log"
	"runtime"

	"github.com/usbarmory/tamago/arm64"
	"github.com/usbarmory/tamago/soc/nxp/enet"
	"github.com/usbarmory/tamago/soc/nxp/imx8mp"

	"github.com/usbarmory/tamago-example/shell"
)

func startInterruptHandler(eth *enet.ENET) {
	imx8mp.GIC.Init()
	imx8mp.GIC.EnableInterrupt(arm64.TIMER_IRQ)

	if eth != nil {
		imx8mp.GIC.EnableInterrupt(eth.IRQ)
	}

	isr := func() {
		irq := imx8mp.GIC.GetInterrupt()

		switch {
		case irq == arm64.TIMER_IRQ:
			imx8mp.ARM64.SetAlarm(0)
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

		imx8mp.ARM64.SetAlarm(pollUntil)
		imx8mp.ARM64.WaitInterrupt()
	}

	arm64.ServiceInterrupts(isr)
}

func Init(console *shell.Interface, hasUSB bool, hasEth bool, nic **enet.ENET) {
	if hasUSB {
		panic("unsupported")
	}

	if hasEth {
		eth := imx8mp.ENET1

		startEth(eth, console, true)
		*nic = eth

		startInterruptHandler(eth)
	}
}
