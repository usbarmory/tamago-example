// https://github.com/usbarmory/tamago-example
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package network

import "net"

const (
	MAC      = "1a:55:89:a2:69:41"
	IP       = "10.0.0.1"
	Resolver = "8.8.8.8:53"
)

func init() {
	net.DefaultNS = []string{Resolver}
}
