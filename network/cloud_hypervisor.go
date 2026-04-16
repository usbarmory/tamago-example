// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build cloud_hypervisor

package network

import (
	"fmt"
	"log"

	"github.com/usbarmory/tamago-example/shell"
	"github.com/usbarmory/tamago/board/cloud_hypervisor/vm"
	"github.com/usbarmory/tamago/kvm/virtio"
	"github.com/usbarmory/tamago/soc/intel/pci"

	"github.com/usbarmory/go-net"
	"github.com/usbarmory/go-net/virtio"
)

// chosen by the application for MSI-X signaling
const VIRTIO_NET0_IRQ = 32

func Init(console *shell.Interface, _ bool, _ bool, nic **vnet.Net) (err error) {
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
		MTU:          gnet.MTU,
	}

	*nic = dev

	if err := dev.Init(); err != nil {
		return fmt.Errorf("could not initialize VirtIO device, %v", err)
	}

	iface, err := initStack(console, dev, true)

	if err != nil {
		return fmt.Errorf("could not start network stack, %v", err)
	}

	dev.Start()

	transport.EnableInterrupt(VIRTIO_NET0_IRQ, vnet.ReceiveQueue)
	startInterruptHandler(dev, iface, vm.AMD64, nil)

	return
}
