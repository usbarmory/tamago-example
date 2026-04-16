// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build mx6ullevk || usbarmory

package network

import (
	"net"

	"github.com/usbarmory/tamago/soc/nxp/usb"

	"github.com/usbarmory/go-net"
	"github.com/usbarmory/go-net/imx-usb"
)

const HostMAC = "1a:55:89:a2:69:42"

func handleUSBInterrupt(usb *usb.USB) {
	usb.ServiceInterrupts()
}

func initEthernetOverUSB(port *usb.USB, stack gnet.Stack) (err error) {
	ecm := &usbnet.ECM{
		Stack: stack,
	}

	ecm.HostMAC, _ = net.ParseMAC(HostMAC)
	ecm.DeviceMAC, _ = net.ParseMAC(MAC)

	if err = ecm.Init(); err != nil {
		return
	}

	port.Device = ecm.Device
	port.Init()
	port.DeviceMode()

	// This example illustrates IRQ handling, alternatively a poller can be
	// used with:
	//   port.Start(device)
	port.EnableInterrupt(usb.IRQ_URI) // reset
	port.EnableInterrupt(usb.IRQ_PCI) // port change detect
	port.EnableInterrupt(usb.IRQ_UI)  // transfer completion

	return
}
