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

	"github.com/f-secure-foundry/tamago/board/nxp/mx6ullevk"
)

func init() {
	cards = append(cards, mx6ullevk.SD1)
	cards = append(cards, mx6ullevk.SD2)
}

func bleConsole(term *terminal.Terminal) (err error) {
	return errors.New("not supported")
}
