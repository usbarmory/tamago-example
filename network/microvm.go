// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build microvm

package network

import (
	"log"

	"github.com/usbarmory/tamago/board/qemu/microvm"
	"github.com/usbarmory/tamago-example/shell"
	"github.com/usbarmory/virtio-net"
)

func Init(handler *shell.Interface, hasUSB bool, hasEth bool, nic **vnet.Net) {
	if hasUSB {
		log.Fatalf("unsupported")
	}

	dev := &vnet.Net{
		Base:         microvm.VIRTIO_NET0_BASE,
		IRQ:          microvm.VIRTIO_NET0_IRQ,
		HeaderLength: 10,
	}

	*nic = dev
	startNet(handler, dev)

	// This example illustrates IRQ handling, alternatively a poller can be
	// used with `dev.Start(true)`.
	dev.Start(false)
	startInterruptHandler(dev, microvm.LAPIC, microvm.IOAPIC1)

	return
}
