// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build firecracker

package network

import (
	"log"

	"github.com/usbarmory/tamago-example/shell"
	"github.com/usbarmory/tamago/board/firecracker/microvm"
	"github.com/usbarmory/tamago/kvm/virtio"

	"github.com/usbarmory/go-net"
	"github.com/usbarmory/go-net/virtio"
)

func Init(console *shell.Interface, hasUSB bool, hasEth bool, nic **vnet.Net) {
	if hasUSB {
		log.Fatalf("unsupported")
	}

	dev := &vnet.Net{
		Transport: &virtio.MMIO{
			Base: microvm.VIRTIO_NET0_BASE,
		},
		IRQ: microvm.VIRTIO_NET0_IRQ,
		MTU: gnet.MTU,
	}

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
		microvm.AMD64.ClearInterrupt()
		dev.Start()
	}()

	startInterruptHandler(dev, iface, microvm.AMD64, microvm.IOAPIC0)
}
