// https://github.com/usbarmory/tamago-example
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build microvm

package network

import (
	"log"
	"net"

	"github.com/usbarmory/tamago/board/qemu/microvm"
	"github.com/usbarmory/virtio-net"
)

const (
	Netmask = "255.255.255.0"
	Gateway = "10.0.0.2"
)

func Init(handler ConsoleHandler, hasUSB bool, hasEth bool) (dev *vnet.Net) {
	if hasUSB {
		log.Fatalf("unsupported")
	}

	dev = &vnet.Net{
		Base: microvm.VIRTIO_NET_BASE,
	}

	iface := vnet.Interface{}

	if err := iface.Init(dev, IP, Netmask, Gateway); err != nil {
		log.Fatalf("could not initialize VirtIO networking, %v", err)
	}

	iface.EnableICMP()

	if handler != nil {
		listenerSSH, err := iface.ListenerTCP4(22)

		if err != nil {
			log.Fatalf("could not initialize SSH listener, %v", err)
		}

		StartSSHServer(listenerSSH, handler)
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

	dev.Start(true)

	// hook interface into Go runtime
	net.SocketFunc = iface.Socket

	return
}
