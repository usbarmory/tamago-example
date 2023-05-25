// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build mx6ullevk || usbarmory
// +build mx6ullevk usbarmory

package cmd

import (
	"bytes"
	"crypto/aes"
	"fmt"
	"log"
	"strings"

	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

const testVectorDCP = "\x75\xf9\x02\x2d\x5a\x86\x7a\xd4\x30\x44\x0f\xee\xc6\x61\x1f\x0a"

func init() {
	if imx6ul.Native && imx6ul.DCP != nil {
		imx6ul.DCP.Init()
	}
}

func testHashDCP() (err error) {
	// NIST.3 test vector
	sum, err := imx6ul.DCP.Sum256(bytes.Repeat([]byte("a"), 1000000))

	if err != nil {
		return
	}

	if bytes.Compare(sum[:], []byte(testVectorNIST3)) != 0 {
		return fmt.Errorf("sum:%x != testVector:%x", sum, testVectorNIST3)
	}

	log.Printf("imx6_dcp: NIST.3 SHA256 %x", sum)

	return
}

func testCipherDCP(keySize int) (err error) {
	buf := make([]byte, aes.BlockSize)
	key := make([]byte, keySize/8)
	iv := make([]byte, aes.BlockSize)

	_ = imx6ul.DCP.SetKey(keySlot, key)

	if err = imx6ul.DCP.Encrypt(buf, keySlot, iv); err != nil {
		return
	}

	if bytes.Compare(buf, []byte(testVector[keySize])) != 0 {
		return fmt.Errorf("buf:%x != testVector:%x", buf, testVector[keySize])
	}

	log.Printf("imx6_dcp: NIST aes-128 cbc encrypt %x", buf)

	if err = imx6ul.DCP.Decrypt(buf, keySlot, iv); err != nil {
		return
	}

	if bytes.Compare(buf, make([]byte, aes.BlockSize)) != 0 {
		return fmt.Errorf("decrypt mismatch (%x)", buf)
	}

	log.Printf("imx6_dcp: NIST aes-128 cbc decrypt %x", buf)

	return
}

func testKeyDerivationDCP() (err error) {
	iv := make([]byte, aes.BlockSize)

	key, err := imx6ul.DCP.DeriveKey([]byte(testDiversifier), iv, -1)

	if err != nil {
		return
	}

	if bytes.Compare(key, make([]byte, len(key))) == 0 {
		return fmt.Errorf("derivedKey all zeros")
	}

	// if the SoC is secure booted we can only print the result
	if imx6ul.HAB() {
		log.Printf("imx6_dcp: derived SNVS key %x", key)
		return
	}

	if strings.Compare(string(key), testVectorDCP) != 0 {
		return fmt.Errorf("derivedKey:%x != testVector:%x", key, testVectorDCP)
	}

	log.Printf("imx6_dcp: derived test key %x", key)

	return
}

func dcpTest() {
	msg("imx6_dcp")

	if !(imx6ul.Native && imx6ul.DCP != nil) {
		log.Printf("skipping imx6_dcp tests under emulation or unsupported hardware")
		return
	}

	if err := testHashDCP(); err != nil {
		log.Printf("imx6_dcp: hash error, %v", err)
	}

	if err := testCipherDCP(128); err != nil {
		log.Printf("imx6_dcp: cipher error, %v", err)
	}

	// derive twice to ensure consistency across repeated operations

	if err := testKeyDerivationDCP(); err != nil {
		log.Printf("imx6_dcp: key derivation error, %v", err)
	}

	if err := testKeyDerivationDCP(); err != nil {
		log.Printf("imx6_dcp: key derivation error, %v", err)
	}
}
