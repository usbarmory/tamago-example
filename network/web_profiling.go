// https://github.com/usbarmory/tamago-example
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package network

import (
	"github.com/arl/statsviz"
	_ "net/http/pprof"
)

func init() {
	statsviz.RegisterDefault()
}
