// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build mx6ullevk || usbarmory
// +build mx6ullevk usbarmory

package cmd

import (
	"crypto/aes"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"

	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

const (
	testVector  = "\x75\xf9\x02\x2d\x5a\x86\x7a\xd4\x30\x44\x0f\xee\xc6\x61\x1f\x0a"
	zeroVector  = "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"
	diversifier = "\xde\xad\xbe\xef"
	keySlot     = 0
)

func init() {
	Add(Cmd{
		Name:    "aes",
		Args:    2,
		Pattern: regexp.MustCompile(`^aes (\d+) (\d+)$`),
		Syntax:  "<size> <sec>",
		Help:    "benchmark DCP hardware encryption",
		Fn:      aesCmd,
	})

	Add(Cmd{
		Name:    "sha",
		Args:    2,
		Pattern: regexp.MustCompile(`^sha (\d+) (\d+)$`),
		Syntax:  "<size> <sec>",
		Help:    "benchmark DCP hardware hashing",
		Fn:      shaCmd,
	})

	if imx6ul.Native && imx6ul.Family == imx6ul.IMX6ULL {
		imx6ul.DCP.Init()
	}
}

func aesCmd(_ *term.Terminal, arg []string) (res string, err error) {
	if !(imx6ul.Native && imx6ul.Family == imx6ul.IMX6ULL) {
		return "", errors.New("unsupported under emulation or incompatible hardware")
	}

	iv := make([]byte, aes.BlockSize)

	if _, err = imx6ul.DCP.DeriveKey([]byte(diversifier), iv, keySlot); err != nil {
		return
	}

	fn := func(b []byte) error {
		return imx6ul.DCP.Decrypt(b, keySlot, iv)
	}

	return dcpCmd(arg, "aes-128 cbc", fn)
}

func shaCmd(_ *term.Terminal, arg []string) (res string, err error) {
	if !(imx6ul.Native && imx6ul.Family == imx6ul.IMX6ULL) {
		return "", errors.New("unsupported under emulation or incompatible hardware")
	}

	fn := func(buf []byte) (err error) {
		_, err = imx6ul.DCP.Sum256(buf)
		return
	}

	return dcpCmd(arg, "sha256", fn)
}

func dcpCmd(arg []string, tag string, fn func(buf []byte) error) (res string, err error) {
	size, err := strconv.Atoi(arg[0])

	if err != nil {
		return "", fmt.Errorf("invalid size, %v", err)
	}

	sec, err := strconv.Atoi(arg[1])

	if err != nil {
		return "", fmt.Errorf("invalid duration, %v", err)
	}

	log.Printf("Doing %s for %ds on %d blocks", tag, sec, size)

	n := 0
	buf := make([]byte, size)
	start := time.Now()

	for run, timeout := true, time.After(time.Duration(sec)*time.Second); run; {
		if err = fn(buf); err != nil {
			return
		}

		n++

		select {
		case <-timeout:
			run = false
		default:
		}
	}

	return fmt.Sprintf("%d %s's in %s", n, tag, time.Since(start)), nil
}

func testKeyDerivation() (err error) {
	iv := make([]byte, aes.BlockSize)

	key, err := imx6ul.DCP.DeriveKey([]byte(diversifier), iv, -1)

	if err != nil {
		return
	}

	if strings.Compare(string(key), zeroVector) == 0 {
		return fmt.Errorf("derivedKey all zeros")
	}

	// if the SoC is secure booted we can only print the result
	if imx6ul.HAB() {
		log.Printf("imx6_dcp: derived SNVS key %x", key)
		return
	}

	if strings.Compare(string(key), testVector) != 0 {
		return fmt.Errorf("derivedKey:%x != testVector:%x", key, testVector)
	}

	log.Printf("imx6_dcp: derived test key %x", key)

	return
}

func dcpTest() {
	msg("imx6_dcp")

	if !(imx6ul.Native && imx6ul.Family == imx6ul.IMX6ULL) {
		log.Printf("skipping imx6_dcp tests under emulation or incompatible hardware")
		return
	}

	// derive twice to ensure consistency across repeated operations

	if err := testKeyDerivation(); err != nil {
		log.Printf("key derivation error, %v", err)
	}

	if err := testKeyDerivation(); err != nil {
		log.Printf("key derivation error, %v", err)
	}
}
