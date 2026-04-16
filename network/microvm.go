// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build microvm

package network

import (
	"fmt"

	"github.com/usbarmory/tamago-example/shell"
	"github.com/usbarmory/tamago/board/qemu/microvm"
	"github.com/usbarmory/tamago/kvm/virtio"

	"github.com/usbarmory/go-net"
	"github.com/usbarmory/go-net/virtio"
)

func Init(console *shell.Interface, _ bool, _ bool, nic **vnet.Net) (err error) {
	dev := &vnet.Net{
		Transport: &virtio.MMIO{
			Base: microvm.VIRTIO_NET0_BASE,
		},
		IRQ:          microvm.VIRTIO_NET0_IRQ,
		HeaderLength: 10,
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
	startInterruptHandler(dev, iface, microvm.AMD64, microvm.IOAPIC1)

	return
}
