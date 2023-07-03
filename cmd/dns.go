// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build mx6ullevk || usbarmory
// +build mx6ullevk usbarmory

package cmd

import (
	"fmt"
	"net"
	"regexp"

	"golang.org/x/term"
)

func init() {
	Add(Cmd{
		Name:    "dns",
		Args:    1,
		Pattern: regexp.MustCompile(`^dns (.*)`),
		Syntax:  "<host>",
		Help:    "resolve domain (requires routing)",
		Fn:      dnsCmd,
	})
}

func dnsCmd(iface *Interface, _ *term.Terminal, arg []string) (res string, err error) {
	cname, err := net.LookupHost(arg[0])

	if err != nil {
		return fmt.Sprintf("query error: %v", err), nil
	}

	return fmt.Sprintf("%+v", cname), nil
}
