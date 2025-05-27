// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build mx6ullevk || usbarmory

package cmd

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"

	"github.com/usbarmory/tamago-example/shell"
	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/soc/nxp/bee"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

func init() {
	shell.Add(shell.Cmd{
		Name:    "bee",
		Args:    2,
		Pattern: regexp.MustCompile(`^bee ([[:xdigit:]]+) ([[:xdigit:]]+)$`),
		Syntax:  "<hex region0> <hex region1>",
		Help:    "BEE OTF AES memory encryption",
		Fn:      beeCmd,
	})

	if imx6ul.Native && imx6ul.BEE != nil {
		imx6ul.BEE.Init()
	}
}

func beeCmd(_ *shell.Interface, arg []string) (res string, err error) {
	if !(imx6ul.Native && imx6ul.BEE != nil) {
		return "", errors.New("unavailable under emulation or unsupported hardware")
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
	imx6ul.ARM.ConfigureMMU(
		bee.AliasRegion0,
		bee.AliasRegion1+bee.AliasRegionSize,
		0,
		arm.MemoryRegion)

	log.Printf("OTF AES 128 CTR encryption enabled:")
	log.Printf("  %#08x-%#08x aliased at %#08x-%#08x",
		region0, region0+bee.AliasRegionSize-1,
		bee.AliasRegion0, bee.AliasRegion0+bee.AliasRegionSize)

	log.Printf("  %#08x-%#08x aliased at %#08x-%#08x",
		region1, region1+bee.AliasRegionSize-1,
		bee.AliasRegion1, bee.AliasRegion1+bee.AliasRegionSize)

	return
}
