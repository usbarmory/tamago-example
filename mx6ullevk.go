// https://github.com/usbarmory/tamago-example
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build mx6ullevk
// +build mx6ullevk

package main

import (
	"errors"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/usbarmory/tamago/board/nxp/mx6ullevk"
	"github.com/usbarmory/tamago/soc/imx6"
)

const boardName = "MCIMX6ULL-EVK"

func init() {
	imx6.I2C1.Init()
	i2c = append(i2c, imx6.I2C1)

	imx6.I2C2.Init()
	i2c = append(i2c, imx6.I2C2)

	cards = append(cards, mx6ullevk.SD1)
	cards = append(cards, mx6ullevk.SD2)
}

func reset() {
	mx6ullevk.Reset()
}

func bleConsole(term *terminal.Terminal) (err error) {
	return errors.New("not supported")
}
