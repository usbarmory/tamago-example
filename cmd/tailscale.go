// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"golang.org/x/term"

	"tailscale.com/tsnet"

	"github.com/usbarmory/tamago-example/network"
)

func init() {
	Add(Cmd{
		Name:    "tailscale",
		Args:    2,
		Pattern: regexp.MustCompile(`^tailscale ([^\s]+)( verbose)?$`),
		Syntax:  "<auth key> (verbose)?",
		Help:    "start network servers on Tailscale tailnet",
		Fn:      tailscaleCmd,
	})
}

func tailscaleCmd(iface *Interface, _ *term.Terminal, arg []string) (res string, err error) {
	s := &tsnet.Server{
		AuthKey:   arg[0],
		Ephemeral: true,
		Hostname:  "tamago",
		Logf:      func(string, ...any) {},
	}

	if len(arg[1]) > 0 {
		s.Logf = func(format string, args ...any) {
			log.Printf("tsnet --- %s", fmt.Sprintf(format, args...))
		}
	}

	status, err := s.Up(context.Background())

	if err != nil {
		return
	}

	ip := status.TailscaleIPs[0].String()
	log.Printf("Tailscale registered IPV4: %s", ip)

	listenerSSH, err := s.Listen("tcp", fmt.Sprintf(":%d", 22))

	if err != nil {
		return
	}

	go network.StartSSHServer(listenerSSH, iface.Start)

	listenerHTTP, err := s.Listen("tcp", fmt.Sprintf(":%d", 80))

	if err != nil {
		return
	}

	listenerHTTPS, err := s.Listen("tcp", fmt.Sprintf(":%d", 443))

	if err != nil {
		return
	}

	go network.StartWebServer(listenerHTTP, status.TailscaleIPs[0].String(), 80, false)
	go network.StartWebServer(listenerHTTPS, status.TailscaleIPs[0].String(), 443, true)

	return
}
