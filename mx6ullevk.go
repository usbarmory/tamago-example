// https://github.com/f-secure-foundry/tamago-example
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// +build mx6ullevk

package main

import (
	"github.com/f-secure-foundry/tamago/board/nxp/mx6ullevk"
)

func init() {
	SD = mx6ullevk.SD
	MMC = mx6ullevk.MMC
}
