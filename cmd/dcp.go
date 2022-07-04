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
	"crypto/aes"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"

	"github.com/usbarmory/tamago/soc/imx6"
	"github.com/usbarmory/tamago/soc/imx6/dcp"
)

const (
	testVector  = "\x75\xf9\x02\x2d\x5a\x86\x7a\xd4\x30\x44\x0f\xee\xc6\x61\x1f\x0a"
	zeroVector  = "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"
	diversifier = "\xde\xad\xbe\xef"
)

func init() {
	Add(Cmd{
		Name: "dcp",
		Args: 2,
		Pattern: regexp.MustCompile(`^dcp (\d+) (\d+)`),
		Syntax: "<size> <sec>",
		Help: "benchmark hardware encryption",
		Fn: dcpCmd,
	})

	if !(imx6.Native && imx6.Family == imx6.IMX6ULL) {
		return
	}

	dcp.Init()
}

func dcpCmd(_ *term.Terminal, arg []string) (res string, err error) {
	if !(imx6.Native && imx6.Family == imx6.IMX6ULL) {
		return "", errors.New("unsupported under emulation")
	}

	size, err := strconv.Atoi(arg[0])

	if err != nil {
		return "", fmt.Errorf("invalid size, %v", err)
	}

	sec, err := strconv.Atoi(arg[1])

	if err != nil {
		return "", fmt.Errorf("invalid duration, %v", err)
	}

	log.Printf("Doing aes-128 cbc for %ds on %d blocks", sec, size)

	n, d, err := testDecryption(size, sec)

	if err != nil {
		return
	}

	return fmt.Sprintf("%d aes-128 cbc's in %s", n, d), nil
}

func testKeyDerivation() (err error) {
	iv := make([]byte, aes.BlockSize)

	key, err := dcp.DeriveKey([]byte(diversifier), iv, -1)

	if err != nil {
		return
	}

	if strings.Compare(string(key), zeroVector) == 0 {
		return fmt.Errorf("derivedKey all zeros")
	}

	// if the SoC is secure booted we can only print the result
	if imx6.SNVS() {
		log.Printf("imx6_dcp: derived SNVS key %x", key)
		return
	}

	// The test vector comparison is left for reference as on non secure
	// booted units it is never reached, as the earlier DeriveKey()
	// invocation returns an error to ensure that no key is derived with
	// public test vectors.
	//
	// Therefore to get here for testing purposes the imx6 package needs to
	// be manually modified to skip the SNVS() check within DeriveKey().

	if strings.Compare(string(key), testVector) != 0 {
		return fmt.Errorf("derivedKey:%x != testVector:%x", key, testVector)
	}

	log.Printf("imx6_dcp: derived test key %x", key)

	return
}

func testDecryption(size int, sec int) (n int, d time.Duration, err error) {
	iv := make([]byte, aes.BlockSize)
	buf := make([]byte, size)

	_, err = dcp.DeriveKey([]byte(diversifier), iv, 0)

	if err != nil {
		return
	}

	start := time.Now()

	for run, timeout := true, time.After(time.Duration(sec)*time.Second); run; {
		if err = dcp.Decrypt(buf, 0, iv); err != nil {
			return
		}

		n++

		select {
		case <-timeout:
			run = false
		default:
		}
	}

	return n, time.Since(start), err
}

func dcpTest() {
	msg("imx6_dcp")

	if !(imx6.Native && imx6.Family == imx6.IMX6ULL) {
		log.Printf("skipping imx6_dcp tests under emulation")
		return
	}

	// derive twice to ensure consistency across repeated operations

	if err := testKeyDerivation(); err != nil {
		log.Printf("imx6_dcp: error, %v", err)
	}

	if err := testKeyDerivation(); err != nil {
		log.Printf("imx6_dcp: error, %v", err)
	}
}
