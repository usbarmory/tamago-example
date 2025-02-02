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
	"github.com/usbarmory/virtio-net"
)

func startInterruptHandler(dev *vnet.Net, lapic *apic.LAPIC, ioapic *apic.IOAPIC) {
	lapic.Enable()

	if dev != nil {
		ioapic.EnableInterrupt(dev.IRQ, dev.IRQ)
	}

	isr := func(irq int) {
		switch {
		case dev != nil && irq == dev.IRQ:
			for buf := dev.Rx(); buf != nil; buf = dev.Rx() {
				dev.RxHandler(buf)
			}
			lapic.ClearInterrupt()
		default:
			log.Printf("internal error, unexpected IRQ %d", irq)
		}
	}

	amd64.ServiceInterrupts(isr)
}

func startNet(handler ConsoleHandler, dev *vnet.Net) {
	iface := vnet.Interface{}

	if err := iface.Init(dev, IP, Netmask, Gateway); err != nil {
		log.Fatalf("could not initialize VirtIO networking, %v", err)
	}

	iface.EnableICMP()

	if handler != nil {
		listenerSSH, err := iface.ListenerTCP4(22)

		if err != nil {
			log.Fatalf("could not initialize SSH listener, %v", err)
		}

		StartSSHServer(listenerSSH, handler)
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
