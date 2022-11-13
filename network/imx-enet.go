// https://github.com/usbarmory/tamago-example
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build mx6ullevk
// +build mx6ullevk

package network

import (
	"log"
	"os"

	"github.com/usbarmory/imx-enet"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

const Gateway = "10.0.0.2"

var (
	iface *enet.Interface
	journal *os.File
)

func Start(console consoleHandler, journalFile *os.File) {
	var err error

	iface, err = enet.Init(imx6ul.ENET1, IP, MAC, Gateway, 1)

	if err != nil {
		log.Fatalf("could not initialize Ethernet networking, %v", err)
	}

	iface.EnableICMP()

	if console != nil {
		listenerSSH, err := iface.ListenerTCP4(22)

		if err != nil {
			log.Fatalf("could not initialize SSH listener, %v", err)
		}

		go func() {
			startSSHServer(listenerSSH, IP, 22, console)
		}()
	}

	listenerHTTP, err := iface.ListenerTCP4(80)

	if err != nil {
		log.Fatalf("could not initialize HTTP listener, %v", err)
	}

	listenerHTTPS, err := iface.ListenerTCP4(443)

	if err != nil {
		log.Fatalf("could not initialize HTTP listener, %v", err)
	}

	// create index.html
	setupStaticWebAssets()

	go func() {
		startWebServer(listenerHTTP, IP, 80, false)
	}()

	go func() {
		startWebServer(listenerHTTPS, IP, 443, true)
	}()

	journal = journalFile

	// never returns
	iface.NIC.Device.Start()
}
