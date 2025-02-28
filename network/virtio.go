// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build firecracker || microvm

package network

import (
	"log"
	"net"

	"github.com/usbarmory/tamago/amd64"
	"github.com/usbarmory/tamago/soc/intel/apic"
	"github.com/usbarmory/tamago-example/shell"
	"github.com/usbarmory/virtio-net"
)

// redirection vector for IOAPIC IRQ to CPU IRQ
const vector = 32

func startInterruptHandler(dev *vnet.Net, lapic *apic.LAPIC, ioapic *apic.IOAPIC) {
	if dev == nil {
		return
	}

	lapic.Enable()
	ioapic.EnableInterrupt(dev.IRQ, vector)

	isr := func(irq int) {
		switch irq {
		case vector:
			for buf := dev.Rx(); buf != nil; buf = dev.Rx() {
				dev.RxHandler(buf)
			}
			lapic.ClearInterrupt()
		case 6:
			// Firecracker bug (?), raised once on
			//   `xorps  %xmm15,%xmm15`
			// which is valid as SSE is enabled.
		default:
			log.Printf("internal error, unexpected IRQ %d", irq)
		}
	}

	amd64.ServiceInterrupts(isr)
}

func startNet(console *shell.Interface, dev *vnet.Net) {
	iface := vnet.Interface{}

	if err := iface.Init(dev, IP, Netmask, Gateway); err != nil {
		log.Fatalf("could not initialize VirtIO networking, %v", err)
	}

	iface.EnableICMP()

	if console != nil {
		listenerSSH, err := iface.ListenerTCP4(22)

		if err != nil {
			log.Fatalf("could not initialize SSH listener, %v", err)
		}

		StartSSHServer(listenerSSH, console)
	}

	listenerHTTP, err := iface.ListenerTCP4(80)

	if err != nil {
		log.Fatalf("could not initialize HTTP listener, %v", err)
	}

	listenerHTTPS, err := iface.ListenerTCP4(443)

	if err != nil {
		log.Fatalf("could not initialize HTTP listener, %v", err)
	}

	StartWebServer(listenerHTTP, IP, 80, false)
	StartWebServer(listenerHTTPS, IP, 443, true)

	// hook interface into Go runtime
	net.SocketFunc = iface.Socket

	return
}
