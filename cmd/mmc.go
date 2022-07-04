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
	"log"
	"regexp"
	"strconv"
	"time"

	"golang.org/x/term"

	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/soc/imx6"
	"github.com/usbarmory/tamago/soc/imx6/usb"
	"github.com/usbarmory/tamago/soc/imx6/usdhc"
)

const (
	// We could use the entire iRAM before USB activation,
	// accounting for required dTD alignment which takes
	// additional space (readSize = 0x20000 - 4096).
	readSize = 0x7fff
	totalReadSize = 10 * 1024 * 1024
)

var MMC []*usdhc.USDHC

func init() {
	Add(Cmd{
		Name: "mmc",
		Args: 3,
		Pattern: regexp.MustCompile(`^mmc (\d) ([[:xdigit:]]+) (\d+)`),
		Syntax: "<n> <hex offset> <size>",
		Help: "MMC/SD card read",
		Fn: mmcCmd,
	})
}

func mmcCmd(_ *term.Terminal, arg []string) (res string, err error) {
	n, err := strconv.ParseUint(arg[0], 10, 8)

	if err != nil {
		return "", fmt.Errorf("invalid card index: %v", err)
	}

	addr, err := strconv.ParseUint(arg[1], 16, 32)

	if err != nil {
		return "", fmt.Errorf("invalid address: %v", err)
	}

	size, err := strconv.ParseUint(arg[2], 10, 32)

	if err != nil {
		return "", fmt.Errorf("invalid size: %v", err)
	}

	if size > maxBufferSize {
		return "", fmt.Errorf("size argument must be <= %d", maxBufferSize)
	}

	if len(MMC) < int(n+1) {
		return "", fmt.Errorf("invalid card index")
	}

	card := MMC[n]

	if err = card.Detect(); err != nil {
		return
	}

	buf, err := card.Read(int64(addr), int64(size))

	if err != nil {
		return
	}

	return hex.Dump(buf), nil
}

func mmcRead(card *usdhc.USDHC, size int, readSize int) {
	if err := card.Detect(); err != nil {
		log.Printf("card error, %v", err)
		return
	}

	info := card.Info()
	capacity := int64(info.BlockSize) * int64(info.Blocks)
	blocks := readSize / info.BlockSize

	giga := capacity / (1000 * 1000 * 1000)
	gibi := capacity / (1024 * 1024 * 1024)

	log.Printf("%d GB/%d GiB card detected %+v", giga, gibi, info)

	addr, buf := dma.Reserve(blocks*info.BlockSize, usb.DTD_PAGE_SIZE)
	defer dma.Release(addr)

	var lba int

	start := time.Now()

	for lba = 0; lba < (size / info.BlockSize); lba += blocks {
		err := card.ReadBlocks(lba, buf)

		if err != nil {
			log.Printf("card read error, %v", err)
			return
		}
	}

	elapsed := time.Since(start)

	// adjust number of read bytes
	size = lba * info.BlockSize

	megaps := (float64(size) / (1000 * 1000)) / elapsed.Seconds()
	mebips := (float64(size) / (1024 * 1024)) / elapsed.Seconds()

	log.Printf("read %d MiB in %s (%.2f MB/s | %.2f MiB/s)", size/(1024*1024), elapsed, megaps, mebips)
}

func mmcTest() {
	msg("imx6_usdhc")

	if !imx6.Native {
		log.Printf("skipping imx6_usdhc tests under emulation")
	}

	for _, card := range MMC {
		mmcRead(card, totalReadSize, readSize)
	}
}
