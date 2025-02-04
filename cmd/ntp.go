// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build mx6ullevk || usbarmory || microvm || firecracker

package cmd

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"time"

	"golang.org/x/term"

	"github.com/beevik/ntp"
)

func init() {
	Add(Cmd{
		Name:    "ntp",
		Args:    1,
		Pattern: regexp.MustCompile(`^ntp (.*)`),
		Syntax:  "<host>",
		Help:    "change runtime date and time via NTP",
		Fn:      ntpCmd,
	})
}

func ntpCmd(iface *Interface, _ *term.Terminal, arg []string) (res string, err error) {
	ip, err := net.DefaultResolver.LookupIP(context.Background(), "ip4", arg[0])

	if err != nil {
		return
	}

	ntpR, err := ntp.QueryWithOptions(
		ip[0].String(),
		ntp.QueryOptions{},
	)

	if err != nil {
		return "", fmt.Errorf("query error: %v", err)
	}

	if err := ntpR.Validate(); err != nil {
		return "", fmt.Errorf("validation error, %v", err)
	}

	date(ntpR.Time.UnixNano())

	return fmt.Sprintf("%s", time.Now().Format(time.RFC3339)), nil
}
