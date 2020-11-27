// https://github.com/f-secure-foundry/tamago-example
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"crypto/aes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/f-secure-foundry/tamago/soc/imx6"
	"github.com/f-secure-foundry/tamago/soc/imx6/dcp"
)

const testVector = "\x75\xf9\x02\x2d\x5a\x86\x7a\xd4\x30\x44\x0f\xee\xc6\x61\x1f\x0a"
const zeroVector = "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"
const diversifier = "\xde\xad\xbe\xef"

func testKeyDerivation() (err error) {
	iv := make([]byte, aes.BlockSize)

	key, err := dcp.DeriveKey([]byte(diversifier), iv, -1)

	if err != nil {
		return
	}

	if strings.Compare(string(key), zeroVector) == 0 {
		err = fmt.Errorf("derivedKey all zeros")
		return
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
		err = fmt.Errorf("derivedKey:%x != testVector:%x", key, testVector)
		return
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
		err = dcp.Decrypt(buf, 0, iv)

		if err != nil {
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

func TestDCP() {
	// derive twice to ensure consistency across repeated operations

	if err := testKeyDerivation(); err != nil {
		log.Printf("imx6_dcp: error, %v", err)
	}

	if err := testKeyDerivation(); err != nil {
		log.Printf("imx6_dcp: error, %v", err)
	}
}
