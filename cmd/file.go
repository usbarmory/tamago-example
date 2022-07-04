// https://github.com/usbarmory/tamago-example
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

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

func devTest() {
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

func fileTest() {
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

	log.Printf("writing %d bytes to %s", len(Banner), path)
	_, err = file.WriteString(Banner)

	if err != nil {
		log.Fatal(err)
	}
	file.Close()

	read, err := ioutil.ReadFile(path)

	if err != nil {
		log.Fatal(err)
	}

	if strings.Compare(Banner, string(read)) != 0 {
		log.Fatal("file comparison mismatch!")
	} else {
		log.Printf("read %s (%d bytes)", path, len(read))
	}

	ls("/dir")
}

func fsTest() {
	msg("fs")
	devTest()
	fileTest()
}
