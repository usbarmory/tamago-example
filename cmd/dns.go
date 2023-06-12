// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build mx6ullevk || usbarmory
// +build mx6ullevk usbarmory

package cmd

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/miekg/dns"
	"golang.org/x/term"
)

var Resolver string

func init() {
	Add(Cmd{
		Name:    "dns",
		Args:    1,
		Pattern: regexp.MustCompile(`^dns (.*)`),
		Syntax:  "<fqdn>",
		Help:    "resolve domain (requires routing)",
		Fn:      dnsCmd,
	})
}

func (iface *Interface) resolve(s string) (r *dns.Msg, rtt time.Duration, err error) {
	if iface.DialTCP4 == nil {
		err = errors.New("network not available")
		return
	}

	if s[len(s)-1:] != "." {
		s += "."
	}

	msg := new(dns.Msg)
	msg.Id = dns.Id()
	msg.RecursionDesired = true

	msg.Question = make([]dns.Question, 1)
	msg.Question[0] = dns.Question{
		Name:   s,
		Qtype:  dns.TypeANY,
		Qclass: dns.ClassINET,
	}

	conn := new(dns.Conn)

	if conn.Conn, err = iface.DialTCP4(Resolver); err != nil {
		return
	}

	c := new(dns.Client)

	return c.ExchangeWithConn(msg, conn)
}

func dnsCmd(iface *Interface, _ *term.Terminal, arg []string) (res string, err error) {
	r, _, err := iface.resolve(arg[0])

	if err != nil {
		return fmt.Sprintf("query error: %v", err), nil
	}

	return fmt.Sprintf("%+v", r), nil
}
