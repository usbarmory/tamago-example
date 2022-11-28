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
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"

	"golang.org/x/term"

	"github.com/usbarmory/tamago/soc/nxp/bee"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

const regionSize = 0x1fffffff

func init() {
	Add(Cmd{
		Name:    "bee",
		Args:    2,
		Pattern: regexp.MustCompile(`^bee ([[:xdigit:]]+) ([[:xdigit:]]+)`),
		Syntax:  "<hex region0> <hex region1>",
		Help:    "BEE OTF AES memory encryption",
		Fn:      beeCmd,
	})
}

func beeCmd(_ *term.Terminal, arg []string) (res string, err error) {
	if !imx6ul.Native {
		return "", errors.New("unsupported under emulation")
	}

	if model := imx6ul.Model(); model != "i.MX6UL" {
		return "", fmt.Errorf("unsupported on %s", model)
	}

	region0, err := strconv.ParseUint(arg[0], 16, 32)

	if err != nil {
		return "", fmt.Errorf("invalid address, %v", err)
	}

	region1, err := strconv.ParseUint(arg[1], 16, 32)

	if err != nil {
		return "", fmt.Errorf("invalid address, %v", err)
	}

	imx6ul.BEE.Init()

	if err = imx6ul.BEE.Enable(uint32(region0), uint32(region1)); err != nil {
		return
	}

	log.Printf("OTF AES 128 CTR encryption enabled:")
	log.Printf("  %#08x-%#08x aliased at %#08x-%#08x",
		region0, region0 + bee.AliasRegionSize,
		bee.AliasRegion0, bee.AliasRegion0 + bee.AliasRegionSize)

	log.Printf("  %#08x-%#08x aliased at %#08x-%#08x",
		region1, region1 + bee.AliasRegionSize,
		bee.AliasRegion1, bee.AliasRegion1 + bee.AliasRegionSize)

	return
}
