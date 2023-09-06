// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//foobar go:build experimental
//foobar +build experimental

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
		Args:    1,
		Pattern: regexp.MustCompile(`^tailscale ([^\s]+)$`),
		Syntax:  "<auth key>",
		Help:    "start network servers on Tailscale tailnet",
		Fn:      tailscaleCmd,
	})
}

func tailscaleCmd(iface *Interface, _ *term.Terminal, arg []string) (res string, err error) {
	s := &tsnet.Server{
		AuthKey:   arg[0],
		Ephemeral: true,
		Hostname:  "tamago",
	}

	s.Logf = func(format string, args ...any) {
		log.Printf("tsnet --- %s", fmt.Sprintf(format, args...))
	}

	status, err := s.Up(context.Background())

	if err != nil {
		return
	}

	log.Printf("Tailscale IPN network status: %+v", status)

	listenerSSH, err := s.Listen("tcp", fmt.Sprintf(":%d", 22))

	if err != nil {
		return
	}

	go network.StartSSHServer(listenerSSH, iface.Start)

	listenerHTTP, err := s.Listen("tcp", fmt.Sprintf(":%d", 80))

	if err != nil {
		return
	}

	go network.StartWebServer(listenerHTTP, status.TailscaleIPs[0].String(), 80, false)

	return
}
