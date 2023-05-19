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

	"golang.org/x/term"

	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

func init() {
	Add(Cmd{
		Name:    "mkv",
		Help:    "CAAM master key verification",
		Fn:      mkvCmd,
	})
}

func mkvCmd(_ *term.Terminal, arg []string) (res string, err error) {
	if !(imx6ul.Native && imx6ul.CAAM != nil) {
		return "", errors.New("unsupported under emulation or incompatible hardware")
	}

	key, err := imx6ul.CAAM.MasterKeyVerification()

	if err != nil {
		return
	}

	return fmt.Sprintf("BKEK: %x", key), nil
}
