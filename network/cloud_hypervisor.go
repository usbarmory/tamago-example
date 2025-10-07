// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build cloud_hypervisor

package network

import (
	"log"

	"github.com/usbarmory/tamago-example/shell"
	"github.com/usbarmory/tamago/board/cloud_hypervisor/vm"
	"github.com/usbarmory/tamago/kvm/virtio"
	"github.com/usbarmory/tamago/soc/intel/pci"
	"github.com/usbarmory/virtio-net"
)

// chosen by the application for MSI-X signaling
const VIRTIO_NET0_IRQ = 32

func Init(console *shell.Interface, hasUSB bool, hasEth bool, nic **vnet.Net) {
	if hasUSB {
		log.Fatalf("unsupported")
	}

	transport := &virtio.PCI{
		Device: pci.Probe(
			0,
			vm.VIRTIO_NET_PCI_VENDOR,
			vm.VIRTIO_NET_PCI_DEVICE,
		),
	}

	dev := &vnet.Net{
		Transport:    transport,
		IRQ:          VIRTIO_NET0_IRQ,
		HeaderLength: 12,
	}

	*nic = dev

	if err := startNet(console, dev); err != nil {
		log.Printf("could not start networking, %v", err)
		return
	}

	// This example illustrates IRQ handling, alternatively a poller can be
	// used with `dev.Start(true)`.
	dev.Start(false)

	transport.EnableInterrupt(VIRTIO_NET0_IRQ, vnet.ReceiveQueue)
	startInterruptHandler(dev, vm.AMD64, nil)

	return
}
