// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build imx8mpevk

package network

import (
	"log"
	"net"

	imxenet "github.com/usbarmory/imx-enet"
	"github.com/usbarmory/tamago-example/shell"
	"github.com/usbarmory/tamago/soc/nxp/enet"
	"github.com/usbarmory/tamago/soc/nxp/imx8mp"
)

func startEth(console *shell.Interface) (eth *enet.ENET) {
	eth = imx8mp.ENET1
	iface := imxenet.Interface{}

	if err := iface.Init(eth, IP, Netmask, MAC, Gateway); err != nil {
		log.Fatalf("could not initialize Ethernet networking, %v", err)
	}

	iface.EnableICMP()

	if console != nil {
		listenerSSH, err := iface.ListenerTCP4(22)

		if err != nil {
			log.Fatalf("could not initialize SSH listener, %v", err)
		}

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

	// hook interface into Go runtime
	net.SocketFunc = iface.Socket

	// never returns
	eth.Start(true)

	return
}

func Init(console *shell.Interface, hasUSB bool, hasEth bool, nic **enet.ENET) {
	var eth *enet.ENET

	if hasUSB {
		panic("unsupported")
	}

	if hasEth {
		eth = startEth(console)
		*nic = eth
	}
}
