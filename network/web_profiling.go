// Copyright (c) WithSecure Corporation
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
