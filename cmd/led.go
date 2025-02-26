// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build usbarmory

package cmd

import (
	"fmt"
	"regexp"

	"github.com/usbarmory/tamago-example/shell"
	usbarmory "github.com/usbarmory/tamago/board/usbarmory/mk2"
)

func init() {
	leds := "white|blue"

	if model, _ := usbarmory.Model(); model == usbarmory.LAN {
		leds += "|yellow|green"
	}

	shell.Add(shell.Cmd{
		Name:    "led",
		Args:    2,
		Pattern: regexp.MustCompile(fmt.Sprintf("^led (%s) (on|off)$", leds)),
		Syntax:  fmt.Sprintf("(%s) (on|off)", leds),
		Help:    "LED control",
		Fn:      ledCmd,
	})
}

func ledCmd(_ *shell.Interface, arg []string) (res string, err error) {
	var on bool

	if arg[1] == "on" {
		on = true
	}

	usbarmory.LED(arg[0], on)

	return
}
