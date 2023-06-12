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
	"regexp"
	"strconv"

	"golang.org/x/term"

	"github.com/usbarmory/tamago/soc/nxp/enet"
)

// Clause 22 access to Clause 45 MMD registers (802.3-2008)
const (
	// 22.2.4.3.11, MMD access control register (Register 13), 802.3-2008
	REGCR = 13
	// 22.2.4.3.12, MMD access address data register (Register 14), 802.3-2008
	ADDAR = 14
)

// Table 22–9, MMD access control register bit definitions, 802.3-2008
const (
	MMD_FN_ADDR = 0b00
	MMD_FN_DATA = 0b01
)

var ENET *enet.ENET

func init() {
	Add(Cmd{
		Name:    "mii",
		Args:    3,
		Pattern: regexp.MustCompile(`^mii ([[:xdigit:]]+) ([[:xdigit:]]+)(?: )?([[:xdigit:]]+)?`),
		Syntax:  "<hex pa> <hex ra> (hex data)?",
		Help:    "show/change eth PHY standard registers",
		Fn:      miiCmd,
	})

	Add(Cmd{
		Name:    "mmd",
		Args:    4,
		Pattern: regexp.MustCompile(`^mmd ([[:xdigit:]]+) ([[:xdigit:]]+) ([[:xdigit:]]+)(?: )?([[:xdigit:]]+)?`),
		Syntax:  "<hex pa> <hex devad> <hex ra> (hex data)?",
		Help:    "show/change eth PHY extended registers",
		Fn:      mmdCmd,
	})
}

// Clause 22 access to standard management registers (802.3-2008)
func miiCmd(_ *Interface, _ *term.Terminal, arg []string) (res string, err error) {
	if ENET == nil {
		return "", errors.New("MII not available")
	}

	pa, err := strconv.ParseUint(arg[0], 16, 5)

	if err != nil {
		return "", fmt.Errorf("invalid physical address, %v", err)
	}

	ra, err := strconv.ParseUint(arg[1], 16, 5)

	if err != nil {
		return "", fmt.Errorf("invalid register address, %v", err)
	}

	if len(arg[2]) > 0 {
		data, err := strconv.ParseUint(arg[2], 16, 16)

		if err != nil {
			return "", fmt.Errorf("invalid data, %v", err)
		}

		ENET.WritePHYRegister(int(pa), int(ra), uint16(data))
	} else {
		res = fmt.Sprintf("%#x", ENET.ReadPHYRegister(int(pa), int(ra)))
	}

	return
}

// Clause 22 access to Clause 45 MMD registers (802.3-2008)
func mmdCmd(_ *Interface, _ *term.Terminal, arg []string) (res string, err error) {
	if ENET == nil {
		return "", errors.New("MII not available")
	}

	pa, err := strconv.ParseUint(arg[0], 16, 5)

	if err != nil {
		return "", fmt.Errorf("invalid physical address, %v", err)
	}

	devad, err := strconv.ParseUint(arg[1], 16, 5)

	if err != nil {
		return "", fmt.Errorf("invalid device address, %v", err)
	}

	ra, err := strconv.ParseUint(arg[2], 16, 16)

	if err != nil {
		return "", fmt.Errorf("invalid register address, %v", err)
	}

	// set address function
	ENET.WritePHYRegister(int(pa), REGCR, (MMD_FN_ADDR<<14)|(uint16(devad)&0x1f))
	// write address value
	ENET.WritePHYRegister(int(pa), ADDAR, uint16(ra))
	// set data function
	ENET.WritePHYRegister(int(pa), REGCR, (MMD_FN_DATA<<14)|(uint16(devad)&0x1f))

	if len(arg[3]) > 0 {
		data, err := strconv.ParseUint(arg[3], 16, 16)

		if err != nil {
			return "", fmt.Errorf("invalid data, %v", err)
		}

		ENET.WritePHYRegister(int(pa), ADDAR, uint16(data))
	} else {
		res = fmt.Sprintf("%#x", ENET.ReadPHYRegister(int(pa), ADDAR))
	}

	return
}
