// https://github.com/usbarmory/tamago-example
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build sifive_u
// +build sifive_u

package network

import "log"

func Init(_ ConsoleHandler, _ bool, _ bool) (_ any) {
	log.Fatal("unsupported")
	return
}
