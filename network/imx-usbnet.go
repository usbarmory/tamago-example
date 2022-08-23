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
)

var (
	iface *usbnet.Interface
	journal *os.File
)

func Start(console consoleHandler, journalFile *os.File) {
	var err error

	iface, err = usbnet.Init(deviceIP, deviceMAC, hostMAC, 1)

	if err != nil {
		log.Fatalf("could not initialize USB networking, %v", err)
	}

	iface.EnableICMP()

	if console != nil {
		listenerSSH, err := iface.ListenerTCP4(22)

		if err != nil {
			log.Fatalf("could not initialize SSH listener, %v", err)
		}

		go func() {
			startSSHServer(listenerSSH, deviceIP, 22, console)
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
		startWebServer(listenerHTTP, deviceIP, 80, false)
	}()

	go func() {
		startWebServer(listenerHTTPS, deviceIP, 443, true)
	}()

	journal = journalFile

	imx6ul.USB1.Init()
	imx6ul.USB1.DeviceMode()
	imx6ul.USB1.Reset()

	// never returns
	imx6ul.USB1.Start(iface.Device())
}
