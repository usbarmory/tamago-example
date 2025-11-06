// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build imx8mpevk || mx6ullevk || usbarmory

package cmd

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"runtime"
	"strconv"

	"github.com/usbarmory/tamago-example/shell"
	"github.com/usbarmory/tamago/soc/nxp/caam"
)

func init() {
	shell.Add(shell.Cmd{
		Name:    "rtic",
		Args:    2,
		Pattern: regexp.MustCompile(`^rtic(?: )?([[:xdigit:]]+)?(?: )?([[:xdigit:]]+)?$`),
		Syntax:  "(<hex start> <hex end>)?",
		Help:    "start RTIC on .text and optional region",
		Fn:      rticCmd,
	})
}

func rticCmd(_ *shell.Interface, arg []string) (res string, err error) {
	var blocks []caam.MemoryBlock

	if CAAM == nil {
		return "", errors.New("unavailable")
	}

	if len(arg[0]) > 0 && len(arg[1]) > 0 {
		start, err := strconv.ParseUint(arg[0], 16, 32)

		if err != nil {
			return "", fmt.Errorf("invalid start address, %v", err)
		}

		end, err := strconv.ParseUint(arg[1], 16, 32)

		if err != nil {
			return "", fmt.Errorf("invalid end address, %v", err)
		}

		if (start%4) != 0 || (end%4) != 0 {
			return "", fmt.Errorf("only 32-bit aligned regions are supported")
		}

		blocks = append(blocks, caam.MemoryBlock{
			Address: uint32(start),
			Length:  uint32(end - start),
		})
	}

	textStart, textEnd := runtime.TextRegion()

	blocks = append(blocks, caam.MemoryBlock{
		Address: uint32(textStart),
		Length:  uint32(textEnd - textStart),
	})

	if err = CAAM.EnableRTIC(blocks); err != nil {
		return
	}

	log.Printf("RTIC enabled:")
	log.Printf("        scan rate %d cycles", caam.RTICThrottle)

	for i, block := range blocks {
		log.Printf("  memory block #%d at %#08x-%#08x", i+1, block.Address, block.Address+block.Length)
	}

	return
}
