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

	"github.com/usbarmory/go-net"
	"github.com/usbarmory/go-net/virtio"
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
		MTU:          gnet.MTU,
	}

	// Google Virtual Private Cloud (GCP) - europe-west1
	MAC = "42:01:0a:84:00:02"
	IP = "10.132.0.2"
	Gateway = "10.132.0.1"

	*nic = dev

	if err := dev.Init(); err != nil {
		log.Printf("could not initialize VirtIO device, %v", err)
		return
	}

	iface, err := initStack(console, dev)

	if err != nil {
		log.Printf("could not start network stack, %v", err)
		return
	}

	go func() {
		// ensure ISR is running before starting the interface
		gcp.AMD64.ClearInterrupt()
		dev.Start()
	}()

	transport.EnableInterrupt(VIRTIO_NET0_IRQ, vnet.ReceiveQueue)
	startInterruptHandler(dev, iface, gcp.AMD64, gcp.IOAPIC0)
}
