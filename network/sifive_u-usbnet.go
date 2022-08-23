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

import (
	"log"
	"os"
)

var journal *os.File

func Start(_ consoleHandler, _ *os.File) {
	log.Fatal("unsupported")
}
