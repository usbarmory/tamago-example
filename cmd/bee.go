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

	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/soc/nxp/bee"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

func init() {
	Add(Cmd{
		Name:    "bee",
		Args:    2,
		Pattern: regexp.MustCompile(`^bee ([[:xdigit:]]+) ([[:xdigit:]]+)$`),
		Syntax:  "<hex region0> <hex region1>",
		Help:    "BEE OTF AES memory encryption",
		Fn:      beeCmd,
	})

	if imx6ul.Native && imx6ul.Family == imx6ul.IMX6UL {
		imx6ul.BEE.Init()
	}
}

func beeCmd(_ *Interface, _ *term.Terminal, arg []string) (res string, err error) {
	if !(imx6ul.Native && imx6ul.Family == imx6ul.IMX6UL) {
		return "", errors.New("unsupported under emulation or unsupported hardware")
	}

	region0, err := strconv.ParseUint(arg[0], 16, 32)

	if err != nil {
		return "", fmt.Errorf("invalid address, %v", err)
	}

	region1, err := strconv.ParseUint(arg[1], 16, 32)

	if err != nil {
		return "", fmt.Errorf("invalid address, %v", err)
	}

	if err = imx6ul.BEE.Enable(uint32(region0), uint32(region1)); err != nil {
		return
	}

	// Caching must be enabled to ensure that BEE hardware limitations
	// concerning access size are respected.
	memAttr := (arm.TTE_AP_001&0b11)<<10 | arm.TTE_CACHEABLE | arm.TTE_BUFFERABLE | arm.TTE_SECTION
	imx6ul.ARM.ConfigureMMU(
		bee.AliasRegion0,
		bee.AliasRegion1+bee.AliasRegionSize,
		memAttr)

	log.Printf("OTF AES 128 CTR encryption enabled:")
	log.Printf("  %#08x-%#08x aliased at %#08x-%#08x",
		region0, region0+bee.AliasRegionSize-1,
		bee.AliasRegion0, bee.AliasRegion0+bee.AliasRegionSize)

	log.Printf("  %#08x-%#08x aliased at %#08x-%#08x",
		region1, region1+bee.AliasRegionSize-1,
		bee.AliasRegion1, bee.AliasRegion1+bee.AliasRegionSize)

	return
}
