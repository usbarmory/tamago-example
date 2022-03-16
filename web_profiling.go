// https://github.com/usbarmory/tamago-example
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"github.com/arl/statsviz"
	_ "net/http/pprof"
)

func init() {
	statsviz.RegisterDefault()
}
