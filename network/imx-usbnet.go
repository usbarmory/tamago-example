// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build mx6ullevk || usbarmory

package network

import (
	"log"
	"net"

	"github.com/usbarmory/imx-usbnet"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
	"github.com/usbarmory/tamago/soc/nxp/usb"
	"github.com/usbarmory/tamago-example/shell"
)

const HostMAC = "1a:55:89:a2:69:42"

func handleUSBInterrupt(usb *usb.USB) {
	usb.ServiceInterrupts()
}

func startUSB(console *shell.Interface) (port *usb.USB) {
	port = imx6ul.USB1

	iface := usbnet.Interface{}

	if err := iface.Init(IP, MAC, HostMAC); err != nil {
		log.Fatalf("could not initialize USB networking, %v", err)
	}

	port.Device = iface.NIC.Device

	iface.EnableICMP()

	if console != nil {
		listenerSSH, err := iface.ListenerTCP4(22)

		if err != nil {
			log.Fatalf("could not initialize SSH listener, %v", err)
		}

		// wait for server to start before responding to USB requests
		StartSSHServer(listenerSSH, console)
	}

	listenerHTTP, err := iface.ListenerTCP4(80)

	if err != nil {
		log.Fatalf("could not initialize HTTP listener, %v", err)
	}

	listenerHTTPS, err := iface.ListenerTCP4(443)

	if err != nil {
		log.Fatalf("could not initialize HTTP listener, %v", err)
	}

	StartWebServer(listenerHTTP, IP, 80, false)
	StartWebServer(listenerHTTPS, IP, 443, true)

	net.SocketFunc = iface.Socket

	port.Init()
	port.DeviceMode()

	// This example illustrates IRQ handling, alternatively a poller can be
	// used with `port.Start(device)`.

	port.EnableInterrupt(usb.IRQ_URI) // reset
	port.EnableInterrupt(usb.IRQ_PCI) // port change detect
	port.EnableInterrupt(usb.IRQ_UI)  // transfer completion

	// hook interface into Go runtime
	net.SocketFunc = iface.Socket

	return
}
