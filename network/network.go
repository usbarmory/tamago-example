// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build imx8mpevk || mx6ullevk || usbarmory || cloud_hypervisor || firecracker || microvm || gcp

package network

import (
	"fmt"
	"log"
	"net"

	// maintained set of TLS roots for any potential TLS client requests
	_ "golang.org/x/crypto/x509roots/fallback"

	"github.com/usbarmory/go-net"
	"github.com/usbarmory/tamago-example/shell"
)

// This example starts TCP/IP networking on all available network
// interfaces (either USB, Ethernet or both), for simplicity each NIC
// is assigned the same IP address and its own gVisor stack.
//
// For more advanced use cases gVisor supports sharing a single stack across
// different NIC IDs and routing while this example simply clones interface
// configuration and stack.
var (
	MAC      = "1a:55:89:a2:69:41"
	Netmask  = "255.255.255.0"
	CIDR     = "/24"
	IP       = "10.0.0.1"
	Gateway  = "10.0.0.2"
	Resolver = "8.8.8.8:53"
)

func initStack(console *shell.Interface, dev gnet.NetworkDevice) (iface *gnet.Interface, err error) {
	iface = &gnet.Interface{}

	if err := iface.Init(dev, IP+CIDR, MAC, Gateway); err != nil {
		return nil, fmt.Errorf("could not initialize stack, %v", err)
	}

	iface.HandleStackErr = func(err error, tx bool) {
		log.Printf("network stack error (tx:%v), %v", tx, err)
	}

	iface.Stack.EnableICMP()

	// hook interface into Go runtime
	net.SetDefaultNS([]string{Resolver})
	net.SocketFunc = iface.Stack.Socket

	if console != nil {
		listenerSSH, err := net.Listen("tcp4", ":22")

		if err != nil {
			return nil, fmt.Errorf("could not initialize SSH listener, %v", err)
		}

		StartSSHServer(listenerSSH, console)
	}

	listenerHTTP, err := net.Listen("tcp4", ":80")

	if err != nil {
		return nil, fmt.Errorf("could not initialize HTTP listener, %v", err)
	}

	listenerHTTPS, err := net.Listen("tcp4", ":443")

	if err != nil {
		return nil, fmt.Errorf("could not initialize HTTP listener, %v", err)
	}

	StartWebServer(listenerHTTP, IP, 80, false)
	StartWebServer(listenerHTTPS, IP, 443, true)

	return
}
