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
	"errors"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/f-secure-foundry/tamago/board/nxp/mx6ullevk"
	"github.com/f-secure-foundry/tamago/soc/imx6"
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
