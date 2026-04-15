// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build cloud_hypervisor || firecracker || microvm || gcp

package network

import (
	"log"
	"runtime/goos"

	"github.com/usbarmory/tamago/amd64"
	"github.com/usbarmory/tamago/soc/intel/ioapic"

	"github.com/usbarmory/go-net"
	"github.com/usbarmory/go-net/virtio"
)

// redirection vector for IOAPIC IRQ to CPU IRQ
const vector = 32

func startInterruptHandler(dev *vnet.Net, iface *gnet.Interface, cpu *amd64.CPU, ioapic *ioapic.IOAPIC) {
	if dev == nil || iface == nil {
		return
	}

	if cpu.LAPIC != nil {
		cpu.LAPIC.Enable()
	}

	if ioapic != nil {
		ioapic.EnableInterrupt(dev.IRQ, vector)
	}

	// as IRQs are enabled, favor slicing dev.ReceiveWithHeader, opposed to
	// dev.Receive for better performance
	size := dev.HeaderLength + gnet.EthernetMaximumSize + gnet.MTU
	buf := make([]byte, size)

	isr := func(irq int) {
		switch irq {
		case vector:
			for {
				if n, err := dev.ReceiveWithHeader(buf); err != nil || n == 0 {
					return
				}

				iface.Stack.RecvInboundPacket(buf[dev.HeaderLength:])
			}
		default:
			log.Printf("internal error, unexpected IRQ %d", irq)
		}
	}

	// optimize CPU idle management as IRQs are enabled
	goos.Idle = func(pollUntil int64) {
		if pollUntil == 0 {
			return
		}

		cpu.SetAlarm(pollUntil)
		cpu.WaitInterrupt()
		cpu.SetAlarm(0)
	}

	cpu.ServiceInterrupts(isr)
}
