// https://github.com/usbarmory/tamago-example
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build mx6ullevk || usbarmory
// +build mx6ullevk usbarmory

package cmd

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"

	"golang.org/x/term"

	"github.com/usbarmory/tamago/soc/nxp/i2c"
)

var I2C []*i2c.I2C

func init() {
	Add(Cmd{
		Name: "i2c",
		Args: 4,
		Pattern: regexp.MustCompile(`^i2c (\d) ([[:xdigit:]]+) ([[:xdigit:]]+) (\d+)`),
		Syntax: "<n> <hex target> <hex addr> <size>",
		Help: "I²C bus read",
		Fn: i2cCmd,
	})
}

func i2cCmd(_ *term.Terminal, arg []string) (res string, err error) {
	n, err := strconv.ParseUint(arg[0], 10, 8)

	if err != nil {
		return "", fmt.Errorf("invalid bus index, %v", err)
	}

	target, err := strconv.ParseUint(arg[1], 16, 7)

	if err != nil {
		return "", fmt.Errorf("invalid target, %v", err)
	}

	addr, err := strconv.ParseUint(arg[2], 16, 32)

	if err != nil {
		return "", fmt.Errorf("invalid address, %v", err)
	}

	size, err := strconv.ParseUint(arg[3], 10, 32)

	if err != nil {
		return "", fmt.Errorf("invalid size, %v", err)
	}

	if size > maxBufferSize {
		return "", fmt.Errorf("size argument must be <= %d", maxBufferSize)
	}

	if n <= 0 || len(I2C) < int(n) {
		return "", fmt.Errorf("invalid bus index")
	}

	buf, err := I2C[n-1].Read(uint8(target), uint32(addr), 1, int(size))

	if err != nil {
		return
	}

	return hex.Dump(buf), nil
}
