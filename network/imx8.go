// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build imx8mpevk

package network

import (
	"github.com/usbarmory/tamago-example/shell"
	"github.com/usbarmory/tamago/soc/nxp/enet"
	"github.com/usbarmory/tamago/soc/nxp/imx8mp"
)

func Init(console *shell.Interface, hasUSB bool, hasEth bool, nic **enet.ENET) {
	if hasUSB {
		panic("unsupported")
	}

	if hasEth {
		eth := imx8mp.ENET1

		startEth(eth, console, false)
		*nic = eth
	}
}
