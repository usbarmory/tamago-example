// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build cloud_hypervisor || firecracker || microvm || gcp

package network

import (
	"fmt"
	"log"
	"net"
	"runtime/goos"

	"github.com/usbarmory/tamago/amd64"
	"github.com/usbarmory/tamago/soc/intel/ioapic"
	"github.com/usbarmory/tamago-example/shell"
	"github.com/usbarmory/virtio-net"
)

// redirection vector for IOAPIC IRQ to CPU IRQ
const vector = 32

func startInterruptHandler(dev *vnet.Net, cpu *amd64.CPU, ioapic *ioapic.IOAPIC) {
	if dev == nil {
		return
	}

	if cpu.LAPIC != nil {
		cpu.LAPIC.Enable()
	}

	if ioapic != nil {
		ioapic.EnableInterrupt(dev.IRQ, vector)
	}

	isr := func(irq int) {
		switch irq {
		case vector:
			for buf := dev.Rx(); buf != nil; buf = dev.Rx() {
				dev.RxHandler(buf)
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

func startNet(console *shell.Interface, dev *vnet.Net) (err error) {
	iface := vnet.Interface{}

	if err := iface.Init(dev, IP, Netmask, Gateway); err != nil {
		return fmt.Errorf("could not initialize VirtIO networking, %v", err)
	}

	iface.EnableICMP()

	if console != nil {
		listenerSSH, err := iface.ListenerTCP4(22)

		if err != nil {
			return fmt.Errorf("could not initialize SSH listener, %v", err)
		}

		StartSSHServer(listenerSSH, console)
	}

	listenerHTTP, err := iface.ListenerTCP4(80)

	if err != nil {
		return fmt.Errorf("could not initialize HTTP listener, %v", err)
	}

	listenerHTTPS, err := iface.ListenerTCP4(443)

	if err != nil {
		return fmt.Errorf("could not initialize HTTP listener, %v", err)
	}

	StartWebServer(listenerHTTP, IP, 80, false)
	StartWebServer(listenerHTTPS, IP, 443, true)

	// hook interface into Go runtime
	net.SocketFunc = iface.Socket

	return
}
