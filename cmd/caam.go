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

const testVectorCAAM = "\xc2\x0c\x77\xec\xad\x89\xdc\x96\xb7\x9f\xc8\xf7\xda\xab\x97\xb4\x2a\xe8\xdf\x98\x3d\x74\x1c\x34\xac\xa8\x63\xca\xeb\x5f\xde\xcd"

func testHashCAAM() (err error) {
	// NIST.3 test vector
	sum256, err := imx6ul.CAAM.Sum256(bytes.Repeat([]byte("a"), 1000000))

	if err != nil {
		return
	}

	if bytes.Compare(sum256[:], []byte(testVectorNIST3)) != 0 {
		return fmt.Errorf("sum256:%x != testVector:%x", sum256, testVectorNIST3)
	}

	log.Printf("imx6_caam: NIST.3 SHA256 %x", sum256)

	return
}

func testCipherCAAM(keySize int) (err error) {
	buf := make([]byte, aes.BlockSize)
	key := make([]byte, keySize/8)
	iv := make([]byte, aes.BlockSize)

	if err = imx6ul.CAAM.Encrypt(buf, key, iv); err != nil {
		return
	}

	if bytes.Compare(buf, []byte(testVector[keySize])) != 0 {
		return fmt.Errorf("buf:%x != testVector:%x", buf, testVector[keySize])
	}

	log.Printf("imx6_caam: NIST aes-%d cbc encrypt %x", keySize, buf)

	if err = imx6ul.CAAM.Decrypt(buf, key, iv); err != nil {
		return
	}

	if bytes.Compare(buf, make([]byte, aes.BlockSize)) != 0 {
		return fmt.Errorf("decrypt mismatch (%x)", buf)
	}

	log.Printf("imx6_caam: NIST aes-%d cbc decrypt %x", keySize, buf)

	return
}

func testKeyDerivationCAAM() (err error) {
	key, err := imx6ul.CAAM.MasterKeyVerification()

	if err != nil {
		return
	}

	if bytes.Compare(key, make([]byte, len(key))) == 0 {
		return fmt.Errorf("derivedKey all zeros")
	}

	// if the SoC is secure booted we can only print the result
	if imx6ul.HAB() {
		log.Printf("imx6_caam: derived MKV key %x", key)
		return
	}

	if strings.Compare(string(key), testVectorCAAM) != 0 {
		return fmt.Errorf("derivedKey:%x != testVector:%x", key, testVectorCAAM)
	}

	log.Printf("imx6_caam: derived test key %x", key)

	return
}

func caamTest() {
	msg("imx6_caam")

	if !(imx6ul.Native && imx6ul.CAAM != nil) {
		log.Printf("skipping imx6_caam tests under emulation or unsupported hardware")
		return
	}

	if err := testHashCAAM(); err != nil {
		log.Printf("imx6_caam: hash error, %v", err)
	}

	for _, n := range []int{128, 192, 256} {
		if err := testCipherCAAM(n); err != nil {
			log.Printf("imx6_caam: cipher error, %v", err)
		}
	}

	// derive twice to ensure consistency across repeated operations

	if err := testKeyDerivationCAAM(); err != nil {
		log.Printf("imx6_caam: key derivation error, %v", err)
	}

	if err := testKeyDerivationCAAM(); err != nil {
		log.Printf("imx6_caam: key derivation error, %v", err)
	}
}
