// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/term"
)

func init() {
	Add(Cmd{
		Name:    "ls",
		Args:    1,
		Pattern: regexp.MustCompile(`^ls(.*)`),
		Syntax:  "(path)?",
		Help:    "list directory contents",
		Fn:      lsCmd,
	})
}

func lsCmd(_ *Interface, term *term.Terminal, arg []string) (string, error) {
	path := strings.TrimSpace(arg[0])

	if len(path) == 0 {
		path = "/"
	}

	return ls(path)
}

func ls(path string) (string, error) {
	var res bytes.Buffer

	fmt.Fprintf(&res, "listing %s:\n", path)

	f, err := os.Open(path)

	if err != nil {
		return "", err
	}

	d, err := f.Stat()

	if err != nil {
		return "", err
	}

	if !d.IsDir() {
		return "", errors.New("expected directory")
	}

	files, err := f.Readdir(-1)

	if err != nil {
		return "", err
	}

	for _, i := range files {
		fmt.Fprintf(&res, " %s (%d bytes)\n", i.Name(), i.Size())
	}

	return res.String(), nil
}

func devTest(log *log.Logger) {
	res, err := ls("/dev")

	log.Print(res)

	if err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, 20)
	path := "/dev/random"

	log.Printf("reading %d bytes from %s", len(buf), path)

	file, err := os.OpenFile(path, os.O_RDONLY, 0600)
	defer file.Close()

	if err != nil {
		log.Fatal(err)
	}

	n, err := file.Read(buf)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("read %s (%d bytes): %x", path, n, buf)
}

func fileTest(log *log.Logger) {
	dirPath := "/dir"
	fileName := "tamago.txt"
	path := filepath.Join(dirPath, fileName)

	if err := os.MkdirAll(dirPath, 0700); err != nil {
		log.Fatal(err)
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	defer file.Close()

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("writing %d bytes to %s", len(Banner), path)

	if _, err = file.WriteString(Banner); err != nil {
		log.Fatal(err)
	}

	read, err := os.ReadFile(path)

	if err != nil {
		log.Fatal(err)
	}

	if strings.Compare(Banner, string(read)) != 0 {
		log.Fatal("file comparison mismatch!")
	} else {
		log.Printf("read %s (%d bytes)", path, len(read))
	}

	res, err := ls("/dir")

	log.Print(res)

	if err != nil {
		log.Fatal(err)
	}
}

func fsTest() (tag string, res string) {
	tag = "fs"

	b := &strings.Builder{}
	log := log.New(b, "", 0)

	devTest(log)
	fileTest(log)

	return tag, b.String()
}
