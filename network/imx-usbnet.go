// https://github.com/usbarmory/tamago-example
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build mx6ullevk || usbarmory
// +build mx6ullevk usbarmory

package network

import (
	"log"
	"os"
	"net"

	"github.com/usbarmory/imx-usbnet"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
	"github.com/usbarmory/tamago/soc/nxp/usb"
)

const hostMAC = "1a:55:89:a2:69:42"

func handleUSBInterrupt(usb *usb.USB) {
	usb.ServiceInterrupts()
}

func StartUSB(console consoleHandler, journalFile *os.File) (port *usb.USB) {
	port = imx6ul.USB1

	iface, err := usbnet.Init(IP, MAC, hostMAC, 1)

	if err != nil {
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

	go StartWebServer(listenerHTTP, IP, 80, false)
	go StartWebServer(listenerHTTPS, IP, 443, true)

	journal = journalFile

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
