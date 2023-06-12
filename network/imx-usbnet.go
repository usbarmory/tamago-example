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

	"github.com/usbarmory/imx-usbnet"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
	"github.com/usbarmory/tamago/soc/nxp/usb"

	"github.com/usbarmory/tamago-example/cmd"
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

		go startSSHServer(listenerSSH, IP, 22, console)
	}

	listenerHTTP, err := iface.ListenerTCP4(80)

	if err != nil {
		log.Fatalf("could not initialize HTTP listener, %v", err)
	}

	listenerHTTPS, err := iface.ListenerTCP4(443)

	if err != nil {
		log.Fatalf("could not initialize HTTP listener, %v", err)
	}

	go startWebServer(listenerHTTP, IP, 80, false)
	go startWebServer(listenerHTTPS, IP, 443, true)

	// 9P is optional. If it can not be started, that's ok.
	listener9P, err := iface.ListenerTCP4(564)

	if err != nil {
		log.Printf("could not initialize 9P listener, %v", err)
	}

	if listener9P != nil {
		// 9p server (see 9p_server.go)
		go func() {
			start9pServer(listener9P, IP, 564, 1)
		}()
	}

	journal = journalFile

	cmd.DialTCP4 = iface.DialTCP4
	cmd.Resolver = Resolver

	port.Init()
	port.DeviceMode()

	// This example illustrates IRQ handling, alternatively a poller can be
	// used with `port.Start(device)`.

	port.EnableInterrupt(usb.IRQ_URI) // reset
	port.EnableInterrupt(usb.IRQ_PCI) // port change detect
	port.EnableInterrupt(usb.IRQ_UI)  // transfer completion

	return
}
