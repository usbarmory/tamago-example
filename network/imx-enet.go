// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build imx8mpevk || mx6ullevk || usbarmory

package network

import (
	"log"
	"net"

	imxenet "github.com/usbarmory/imx-enet"
	"github.com/usbarmory/tamago/soc/nxp/enet"
	"github.com/usbarmory/tamago-example/shell"
)

func handleEthernetInterrupt(eth *enet.ENET) {
	for buf := eth.Rx(); buf != nil; buf = eth.Rx() {
		eth.RxHandler(buf)
		eth.ClearInterrupt(enet.IRQ_RXF)
	}
}

func startEth(eth *enet.ENET, console *shell.Interface, irq bool) {
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

	if irq {
		eth.EnableInterrupt(enet.IRQ_RXF)
	}

	eth.Start(!irq)
}
