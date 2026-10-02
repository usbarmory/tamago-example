// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build arm

package cmd

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/usbarmory/armory-boot/exec"
	"github.com/usbarmory/tamago-example/shell"
	"github.com/usbarmory/tamago/dma"
)

const (
	memoryStart = 0x90000000
	memorySize  = 0x10000000

	kernelOffset = 0x00800000
	paramsOffset = 0x07000000
	initrdOffset = 0x08000000

	commandLine = "console=ttymxc0,115200"
)

func init() {
	shell.Add(shell.Cmd{
		Name:    "linux",
		Args:    3,
		Pattern: regexp.MustCompile(`^linux\s+(\S+)\s+(\S+)\s+(\S+)`),
		Syntax:  "<zImage> <dtb> <initrd>",
		Help:    "boot Linux kernel zImage",
		Fn:      linuxCmd,
	})
}

// For a full ARM bare metal Go bootloader implementation see [armory-boot].
//
// [armory-boot]: https://github.com/usbarmory/armory-boot
func linuxCmd(_ *shell.Interface, arg []string) (res string, err error) {
	zImage, err := os.ReadFile(strings.TrimSpace(arg[0]))

	if err != nil {
		return
	}

	dtb, err := os.ReadFile(strings.TrimSpace(arg[1]))

	if err != nil {
		return
	}

	initrd, err := os.ReadFile(strings.TrimSpace(arg[2]))

	if err != nil {
		return
	}

	mem, err := dma.NewRegion(memoryStart, memorySize, false)

	if err != nil {
		return
	}

	mem.Reserve(memorySize, 0)

	image := &exec.LinuxImage{
		Region:               mem,
		Kernel:               zImage,
		DeviceTreeBlob:       dtb,
		InitialRamDisk:       initrd,
		KernelOffset:         kernelOffset,
		DeviceTreeBlobOffset: paramsOffset,
		InitialRamDiskOffset: initrdOffset,
		CmdLine:              commandLine,
	}

	if err = image.Load(); err != nil {
		return "", fmt.Errorf("could not load kernel, %v", err)
	}

	log.Printf("starting kernel@%0.8x\n", image.Entry())

	return "", image.Boot(nil)
}
