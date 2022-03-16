// https://github.com/usbarmory/tamago-example
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/soc/imx6/usb"
	"github.com/usbarmory/tamago/soc/imx6/usdhc"
)

var cards []*usdhc.USDHC

func TestUSDHC(card *usdhc.USDHC, size int, readSize int) {
	if err := card.Detect(); err != nil {
		log.Printf("imx6_usdhc: error, %v", err)
		return
	}

	info := card.Info()
	capacity := int64(info.BlockSize) * int64(info.Blocks)
	blocks := readSize / info.BlockSize

	giga := capacity / (1000 * 1000 * 1000)
	gibi := capacity / (1024 * 1024 * 1024)

	log.Printf("imx6_usdhc: %d GB/%d GiB card detected %+v", giga, gibi, info)

	addr, buf := dma.Reserve(blocks*info.BlockSize, usb.DTD_PAGE_SIZE)
	defer dma.Release(addr)

	var lba int

	start := time.Now()

	for lba = 0; lba < (size / info.BlockSize); lba += blocks {
		err := card.ReadBlocks(lba, buf)

		if err != nil {
			log.Printf("imx6_usdhc: card read error, %v", err)
			return
		}
	}

	elapsed := time.Since(start)

	// adjust number of read bytes
	size = lba * info.BlockSize

	megaps := (float64(size) / (1000 * 1000)) / elapsed.Seconds()
	mebips := (float64(size) / (1024 * 1024)) / elapsed.Seconds()

	log.Printf("imx6_usdhc: read %d MiB in %s (%.2f MB/s | %.2f MiB/s)", size/(1024*1024), elapsed, megaps, mebips)
}

func TestDev() {
	ls("/dev")

	buf := make([]byte, 20)
	path := "/dev/random"

	log.Printf("reading %d bytes from %s", len(buf), path)
	file, err := os.OpenFile(path, os.O_RDONLY, 0600)

	if err != nil {
		log.Fatal(err)
	}

	n, err := file.Read(buf)

	if err != nil {
		log.Fatal(err)
	}
	file.Close()

	log.Printf("read %s (%d bytes): %x", path, n, buf)
}

func TestFile() {
	dirPath := "/dir"
	fileName := "tamago.txt"
	path := filepath.Join(dirPath, fileName)

	err := os.MkdirAll(dirPath, 0700)

	if err != nil {
		log.Fatal(err)
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("writing %d bytes to %s", len(banner), path)
	_, err = file.WriteString(banner)

	if err != nil {
		log.Fatal(err)
	}
	file.Close()

	read, err := ioutil.ReadFile(path)

	if err != nil {
		log.Fatal(err)
	}

	if strings.Compare(banner, string(read)) != 0 {
		log.Println("TestFile: comparison fail")
	} else {
		log.Printf("read %s (%d bytes)", path, len(read))
	}

	ls("/dir")
}

func ls(path string) {
	log.Printf("listing %s:", path)

	f, err := os.Open(path)

	if err != nil {
		log.Fatal(err)
	}

	d, err := f.Stat()

	if err != nil {
		log.Fatal(err)
	}

	if !d.IsDir() {
		log.Fatal("expected directory")
	}

	files, err := f.Readdir(-1)

	if err != nil {
		log.Fatal(err)
	}

	for _, i := range files {
		log.Printf(" %s/%s (%d bytes)", path, i.Name(), i.Size())
	}
}
