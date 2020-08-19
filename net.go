// https://github.com/f-secure-foundry/tamago-example
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"log"
	"net"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/link/channel"
	"gvisor.dev/gvisor/pkg/tcpip/network/arp"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/icmp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/waiter"
)

const (
	MTU = 1500

	hostMAC   = "1a:55:89:a2:69:42"
	deviceMAC = "1a:55:89:a2:69:41"

	IP = "10.0.0.1"
)

func configureNetworkStack(addr tcpip.Address, nic tcpip.NICID) (s *stack.Stack, link *channel.Endpoint) {
	var err error

	s = stack.New(stack.Options{
		NetworkProtocols: []stack.NetworkProtocol{
			ipv4.NewProtocol(),
			arp.NewProtocol()},
		TransportProtocols: []stack.TransportProtocol{
			tcp.NewProtocol(),
			icmp.NewProtocol4()},
	})

	linkAddr, err := tcpip.ParseMACAddress(deviceMAC)

	if err != nil {
		log.Fatal(err)
	}

	link = channel.New(256, MTU, linkAddr)
	linkEP := stack.LinkEndpoint(link)

	if err := s.CreateNIC(nic, linkEP); err != nil {
		log.Fatal(err)
	}

	if err := s.AddAddress(nic, arp.ProtocolNumber, arp.ProtocolAddress); err != nil {
		log.Fatal(err)
	}

	if err := s.AddAddress(nic, ipv4.ProtocolNumber, addr); err != nil {
		log.Fatal(err)
	}

	subnet, err := tcpip.NewSubnet("\x00\x00\x00\x00", "\x00\x00\x00\x00")

	if err != nil {
		log.Fatal(err)
	}

	s.SetRouteTable([]tcpip.Route{{
		Destination: subnet,
		NIC:         nic,
	}})

	return
}

func startICMPEndpoint(s *stack.Stack, addr tcpip.Address, port uint16, nic tcpip.NICID) {
	var wq waiter.Queue

	fullAddr := tcpip.FullAddress{Addr: addr, Port: port, NIC: nic}
	ep, err := s.NewEndpoint(icmp.ProtocolNumber4, ipv4.ProtocolNumber, &wq)

	if err != nil {
		log.Fatalf("endpoint error (icmp): %v", err)
	}

	if err := ep.Bind(fullAddr); err != nil {
		log.Fatal("bind error (icmp endpoint): ", err)
	}
}

// StartNetworking starts SSH and HTTP services.
func StartNetworking() (link *channel.Endpoint) {
	addr := tcpip.Address(net.ParseIP(IP)).To4()

	s, link := configureNetworkStack(addr, 1)

	// handle pings
	startICMPEndpoint(s, addr, 0, 1)

	// create index.html
	setupStaticWebAssets()

	// HTTP web server (see web_server.go)
	go func() {
		startWebServer(s, addr, 80, 1, false)
	}()

	// HTTPS web server (see web_server.go)
	go func() {
		startWebServer(s, addr, 443, 1, true)
	}()

	// SSH server (see ssh_server.go)
	go func() {
		startSSHServer(s, addr, 22, 1)
	}()

	return
}
