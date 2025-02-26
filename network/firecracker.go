// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build firecracker

package network

import (
	"log"

	"github.com/usbarmory/tamago/board/firecracker/microvm"
	"github.com/usbarmory/tamago-example/shell"
	"github.com/usbarmory/virtio-net"
)

func Init(console *shell.Interface, hasUSB bool, hasEth bool, nic **vnet.Net) {
	if hasUSB {
		log.Fatalf("unsupported")
	}

	dev := &vnet.Net{
		Base: microvm.VIRTIO_NET0_BASE,
		IRQ:  microvm.VIRTIO_NET0_IRQ,
	}

	*nic = dev
	startNet(console, dev)

	// This example illustrates IRQ handling, alternatively a poller can be
	// used with `dev.Start(true)`.
	dev.Start(false)
	startInterruptHandler(dev, microvm.LAPIC, microvm.IOAPIC0)

	return
}
