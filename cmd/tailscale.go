// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"tailscale.com/tsnet"

	"github.com/usbarmory/tamago-example/network"
	"github.com/usbarmory/tamago-example/shell"
)

func init() {
	shell.Add(shell.Cmd{
		Name:    "tailscale",
		Args:    2,
		Pattern: regexp.MustCompile(`^tailscale ([^\s]+)( verbose)?$`),
		Syntax:  "<auth key> (verbose)?",
		Help:    "start network servers on Tailscale tailnet",
		Fn:      tailscaleCmd,
	})
}

func tailscaleCmd(console *shell.Interface, arg []string) (res string, err error) {
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

	c := *console
	network.StartSSHServer(listenerSSH, &c)

	listenerHTTP, err := s.Listen("tcp", fmt.Sprintf(":%d", 80))

	if err != nil {
		return
	}

	listenerHTTPS, err := s.Listen("tcp", fmt.Sprintf(":%d", 443))

	if err != nil {
		return
	}

	network.StartWebServer(listenerHTTP, status.TailscaleIPs[0].String(), 80, false)
	network.StartWebServer(listenerHTTPS, status.TailscaleIPs[0].String(), 443, true)

	return
}
