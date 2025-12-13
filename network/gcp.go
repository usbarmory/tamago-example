// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build gcp

package network

import (
	"log"

	"github.com/usbarmory/tamago-example/shell"
	"github.com/usbarmory/tamago/board/google/gcp"
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

	transport := &virtio.LegacyPCI{
		Device: pci.Probe(
			0,
			gcp.VIRTIO_NET_PCI_VENDOR,
			gcp.VIRTIO_NET_PCI_DEVICE,
		),
	}

	dev := &vnet.Net{
		Transport:    transport,
		IRQ:          VIRTIO_NET0_IRQ,
		HeaderLength: 10,
	}

	MAC      = "42:01:0a:84:00:02"
	IP       = "10.132.0.2"
	Gateway  = "10.132.0.1"

	*nic = dev

	if err := startNet(console, dev); err != nil {
		log.Printf("could not start networking, %v", err)
		return
	}

	// This example illustrates IRQ handling, alternatively a poller can be
	// used with `dev.Start(true)`.
	dev.Start(false)

	transport.EnableInterrupt(VIRTIO_NET0_IRQ, vnet.ReceiveQueue)
	startInterruptHandler(dev, gcp.AMD64, gcp.IOAPIC0)
}
