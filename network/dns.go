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
	"net"
	"regexp"
	"time"

	"github.com/miekg/dns"
	"golang.org/x/term"

	"github.com/usbarmory/tamago-example/cmd"
)

var dialTCP4 func(string) (net.Conn, error)

func init() {
	cmd.Add(cmd.Cmd{
		Name:    "dns",
		Args:    1,
		Pattern: regexp.MustCompile(`^dns (.*)`),
		Syntax:  "<fqdn>",
		Help:    "resolve domain (requires routing)",
		Fn:      dnsCmd,
	})
}

func resolve(s string) (r *dns.Msg, rtt time.Duration, err error) {
	if s[len(s)-1:] != "." {
		s += "."
	}

	msg := new(dns.Msg)
	msg.Id = dns.Id()
	msg.RecursionDesired = true

	msg.Question = make([]dns.Question, 1)
	msg.Question[0] = dns.Question{s, dns.TypeANY, dns.ClassINET}

	conn := new(dns.Conn)

	if conn.Conn, err = dialTCP4(Resolver); err != nil {
		return
	}

	c := new(dns.Client)

	return c.ExchangeWithConn(msg, conn)
}

func dnsCmd(_ *term.Terminal, arg []string) (res string, err error) {
	if dialTCP4 == nil {
		return "", errors.New("network not available")
	}

	r, _, err := resolve(arg[0])

	if err != nil {
		return fmt.Sprintf("query error: %v", err), nil
	}

	return fmt.Sprintf("%+v", r), nil
}
