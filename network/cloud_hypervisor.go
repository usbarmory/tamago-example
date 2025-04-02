// Copyright (c) WithSecure Corporation
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

func Init(console *shell.Interface, hasUSB bool, hasEth bool, nic **vnet.Net) {
	if hasUSB {
		log.Fatalf("unsupported")
	}

	dev := &vnet.Net{
		Transport: &virtio.PCI{
			Device: pci.Probe(
				0,
				vm.VIRTIO_NET_PCI_VENDOR,
				vm.VIRTIO_NET_PCI_DEVICE,
			),
		},
		IRQ:          0, // FIXME
		HeaderLength: 12,
	}

	*nic = dev

	if err := startNet(console, dev); err != nil {
		log.Printf("could not start networking, %v", err)
		return
	}

	// This example illustrates IRQ handling, alternatively a poller can be
	// used with `dev.Start(true)`.
	dev.Start(true)
	//startInterruptHandler(dev, vm.LAPIC, vm.IOAPIC0)

	return
}
