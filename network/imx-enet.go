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

	imxenet "github.com/usbarmory/imx-enet"
	"github.com/usbarmory/tamago/soc/nxp/enet"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"

	"github.com/usbarmory/tamago-example/cmd"
)

const (
	Netmask = "255.255.255.0"
	Gateway = "10.0.0.2"
)

func handleEthernetInterrupt(eth *enet.ENET) {
	for buf := eth.Rx(); buf != nil; buf = eth.Rx() {
		eth.RxHandler(buf)
		eth.ClearInterrupt(enet.IRQ_RXF)
	}
}

func StartEth(console *cmd.Interface, journalFile *os.File) (eth *enet.ENET) {
	eth = imx6ul.ENET2

	if !imx6ul.Native {
		eth = imx6ul.ENET1
	}

	iface, err := imxenet.Init(eth, IP, Netmask, MAC, Gateway, 1)

	if err != nil {
		log.Fatalf("could not initialize Ethernet networking, %v", err)
	}

	iface.EnableICMP()

	if console != nil {
		listenerSSH, err := iface.ListenerTCP4(22)

		if err != nil {
			log.Fatalf("could not initialize SSH listener, %v", err)
		}

		console.DialTCP4 = iface.DialTCP4
		console.ListenerTCP4 = iface.ListenerTCP4

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

	journal = journalFile

	cmd.ENET = iface.NIC.Device
	cmd.Resolver = Resolver

	// This example illustrates IRQ handling, alternatively a poller can be
	// used with `eth.Start(true)`.

	eth.EnableInterrupt(enet.IRQ_RXF)
	eth.Start(false)

	return
}
