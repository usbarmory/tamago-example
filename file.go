// https://github.com/f-secure-foundry/tamago-example
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

	"github.com/f-secure-foundry/tamago/imx6/usdhc"
)

func TestUSDHC(card *usdhc.USDHC, count int, readSize int) {
	err := card.Detect()

	if err != nil {
		log.Printf("imx6_usdhc: card error, %v", err)
	} else {
		info := card.Info()
		capacity := int64(info.BlockSize) * int64(info.Blocks)
		giga := capacity / (1000 * 1000 * 1000)
		gibi := capacity / (1024 * 1024 * 1024)

		log.Printf("imx6_usdhc: %d GB/%d GiB card detected %+v", giga, gibi, info)

		start := time.Now()

		for i := 0; i < count; i += readSize {
			_, err := card.Read(i, readSize)

			if err != nil {
				log.Printf("imx6_usdhc: card read error, %v", err)
				return
			}
		}

		elapsed := time.Since(start)
		megaps := (float64(count) / (1000 * 1000)) / elapsed.Seconds()
		mebips := (float64(count) / (1024 * 1024)) / elapsed.Seconds()

		log.Printf("imx6_usdhc: read %d MiB in %s (%.2f MB/s | %.2f MiB/s)", count/(1024*1024), elapsed, megaps, mebips)
	}
}

func TestFile() {
	var err error

	defer func() {
		if err != nil {
			log.Printf("TestFile error: %v", err)
		}
	}()

	dirPath := "/dir"
	fileName := "tamago.txt"
	path := filepath.Join(dirPath, fileName)

	log.Printf("writing %d bytes to %s", len(banner), path)

	err = os.MkdirAll(dirPath, 0700)

	if err != nil {
		return
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		panic(err)
	}

	_, err = file.WriteString(banner)

	if err != nil {
		panic(err)
	}
	file.Close()

	read, err := ioutil.ReadFile(path)

	if err != nil {
		panic(err)
	}

	if strings.Compare(banner, string(read)) != 0 {
		log.Println("TestFile: comparison fail")
	} else {
		log.Printf("read %s (%d bytes)", path, len(read))
	}
}

func TestDir() {
	dirPath := "/dir"

	log.Printf("listing directory %s", dirPath)

	f, err := os.Open(dirPath)

	if err != nil {
		panic(err)
	}

	d, err := f.Stat()

	if err != nil {
		panic(err)
	}

	if !d.IsDir() {
		panic("expected directory")
	}

	files, err := f.Readdir(-1)

	if err != nil {
		panic(err)
	}

	for _, i := range files {
		log.Printf("%s/%s (%d bytes)", dirPath, i.Name(), i.Size())
	}
}
