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
	"github.com/usbarmory/tamago/soc/intel/pci"
	"github.com/usbarmory/virtio-net"
)

func Init(console *shell.Interface, hasUSB bool, hasEth bool, nic **vnet.Net) {
	var dev *vnet.Net

	if hasUSB {
		log.Fatalf("unsupported")
	}

	if p := pci.Probe(0, vm.VIRTIO_NET_PCI_VENDOR, vm.VIRTIO_NET_PCI_DEVICE); p != nil {
		dev := &vnet.Net{
			Base:         p.BaseAddress0,
			IRQ:          0, // FIXME
			HeaderLength: 10,
		}
		log.Printf("found VirtIO network %04x:%04x %#x", p.Vendor, p.Device, dev.Base)
		return // WiP
	} else {
		log.Printf("could not find network device")
		return
	}

	*nic = dev
	startNet(console, dev)

	// This example illustrates IRQ handling, alternatively a poller can be
	// used with `dev.Start(true)`.
	dev.Start(false)
	startInterruptHandler(dev, vm.LAPIC, vm.IOAPIC0)

	return
}
