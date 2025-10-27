// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build imx8mpevk

package cmd

import (
	"errors"

	"github.com/usbarmory/tamago-example/shell"
	"github.com/usbarmory/tamago/board/nxp/imx8mpevk"
)

const boardName = "8MPLUSLPD4-EVK"

func init() {
	Terminal = imx8mpevk.UART1
}

func rebootCmd(_ *shell.Interface, _ []string) (_ string, err error) {
	return "", errors.New("unimplemented")
}

func HasNetwork() (usb bool, eth bool) {
	return false, false
}
