// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build usbarmory
// +build usbarmory

package cmd

import (
	"fmt"
	"regexp"

	"golang.org/x/term"

	usbarmory "github.com/usbarmory/tamago/board/usbarmory/mk2"
)

func init() {
	leds := "white|blue"

	if usbarmory.Model() == "UA-MKII-LAN" {
		leds += "|yellow|green"
	}

	Add(Cmd{
		Name:    "led",
		Args:    2,
		Pattern: regexp.MustCompile(fmt.Sprintf("^led (%s) (on|off)$", leds)),
		Syntax:  fmt.Sprintf("(%s) (on|off)", leds),
		Help:    "LED control",
		Fn:      ledCmd,
	})
}

func ledCmd(_ *Interface, _ *term.Terminal, arg []string) (res string, err error) {
	var on bool

	if arg[1] == "on" {
		on = true
	}

	usbarmory.LED(arg[0], on)

	return
}
