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
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"

	"golang.org/x/term"

	imxenet "github.com/usbarmory/imx-enet"
	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/soc/nxp/enet"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"

	"github.com/usbarmory/tamago-example/cmd"
)

const (
	Netmask = "255.255.255.0"
	Gateway = "10.0.0.2"
)

var miiDevice *enet.ENET

func init() {
	cmd.Add(cmd.Cmd{
		Name:    "mii",
		Args:    3,
		Pattern: regexp.MustCompile(`^mii ([[:xdigit:]]+) ([[:xdigit:]]+)(?: )?([[:xdigit:]]+)?`),
		Syntax:  "<hex pa> <hex ra> (hex data)?",
		Help:    "Ethernet IEEE 802.3 MII access",
		Fn:      miiCmd,
	})
}

func miiCmd(_ *term.Terminal, arg []string) (res string, err error) {
	if miiDevice == nil {
		return "", errors.New("MII device not available")
	}

	pa, err := strconv.ParseUint(arg[0], 16, 5)

	if err != nil {
		return "", fmt.Errorf("invalid physical address, %v", err)
	}

	ra, err := strconv.ParseUint(arg[1], 16, 5)

	if err != nil {
		return "", fmt.Errorf("invalid address, %v", err)
	}

	if len(arg[2]) > 0 {
		data, err := strconv.ParseUint(arg[2], 16, 16)

		if err != nil {
			return "", fmt.Errorf("invalid data, %v", err)
		}

		miiDevice.WriteMII(int(pa), int(ra), uint16(data))
	} else {
		res = fmt.Sprintf("%#x", miiDevice.ReadMII(int(pa), int(ra)))
	}

	return
}

func handleInterrupt(eth *enet.ENET) {
	irq, end := imx6ul.GIC.GetInterrupt(true)

	if end != nil {
		end <- true
	}

	if irq != eth.IRQ {
		log.Printf("internal error, unexpected IRQ %d", irq)
		return
	}

	for buf := eth.Rx(); buf != nil; buf = eth.Rx() {
		eth.RxHandler(buf)
		eth.ClearInterrupt(enet.IRQ_RXF)
	}
}

func startInterface(eth *enet.ENET) {
	imx6ul.GIC.Init(true, false)
	imx6ul.GIC.EnableInterrupt(eth.IRQ, true)

	eth.EnableInterrupt(enet.IRQ_RXF)
	eth.Start(false)

	arm.RegisterInterruptHandler()

	for {
		arm.WaitInterrupt()
		handleInterrupt(eth)
	}
}

func StartEth(console consoleHandler, journalFile *os.File) {
	nic := imx6ul.ENET2

	if !imx6ul.Native {
		nic = imx6ul.ENET1
	}

	iface, err := imxenet.Init(nic, IP, Netmask, MAC, Gateway, 1)

	if err != nil {
		log.Fatalf("could not initialize Ethernet networking, %v", err)
	}

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

	journal = journalFile
	dialTCP4 = iface.DialTCP4
	miiDevice = iface.NIC.Device

	// never returns
	startInterface(iface.NIC.Device)
}
