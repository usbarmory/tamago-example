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
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"strings"

	"github.com/dustinxie/ecc"

	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

const testVectorCAAM = "\x49\x3f\xb1\xe8\x7a\x39\x3f\x47\xe8\x3a\xc8\xa2\x27\xfd\x5b\x71\x92\x87\xdb\xad\x13\x2a\x9a\x8d\x8e\xe9\xbd\x3f\x76\x16\x7f\xcb"

func init() {
	if imx6ul.CAAM != nil {
		imx6ul.CAAM.DeriveKeyMemory, _ = dma.NewRegion(imx6ul.OCRAM_START, imx6ul.OCRAM_SIZE, false)
	}
}

func testHashCAAM(log *log.Logger) (err error) {
	sum256, err := imx6ul.CAAM.Sum256([]byte(testVectorSHAInput))

	if err != nil {
		return
	}

	if bytes.Compare(sum256[:], []byte(testVectorSHA)) != 0 {
		return fmt.Errorf("sum256:%x != testVector:%x", sum256, testVectorSHA)
	}

	log.Printf("imx6_caam: FIPS 180-2 SHA256 %x", sum256)

	return
}

func testCipherCAAM(keySize int, log *log.Logger) (err error) {
	buf := bytes.Clone([]byte(testVectorInput))
	key := []byte(testVectorKey[keySize])
	iv := []byte(testVectorIV)

	if err = imx6ul.CAAM.Encrypt(buf, key, iv); err != nil {
		return
	}

	if bytes.Compare(buf, []byte(testVectorCipher[keySize])) != 0 {
		return fmt.Errorf("buf:%x != testVector:%x", buf, testVectorCipher[keySize])
	}

	log.Printf("imx6_caam: NIST aes-%d cbc encrypt %x", keySize, buf)

	if err = imx6ul.CAAM.Decrypt(buf, key, iv); err != nil {
		return
	}

	if bytes.Compare(buf, []byte(testVectorInput)) != 0 {
		return fmt.Errorf("decrypt mismatch (%x)", buf)
	}

	log.Printf("imx6_caam: NIST aes-%d cbc decrypt %x", keySize, buf)

	cmac, err := imx6ul.CAAM.SumAES([]byte(testVectorInput), key)

	if err != nil {
		return
	}

	if bytes.Compare(cmac[:], []byte(testVectorMAC[keySize])) != 0 {
		return fmt.Errorf("cmac:%x != testVector:%x", cmac, testVectorMAC[keySize])
	}

	log.Printf("imx6_caam: NIST.3 aes-%d cmac %x", keySize, cmac)

	return
}

func testSignatureCAAM(log *log.Logger) (err error) {
	hash := make([]byte, sha256.Size)
	_, _ = rand.Read(hash)

	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	r, s, err := imx6ul.CAAM.Sign(priv, hash, nil)

	if err != nil {
		log.Printf("%v", err)
		return err
	}

	if !ecdsa.Verify(&priv.PublicKey, hash, r, s) {
		return fmt.Errorf("invalid ecdsap256 signature")
	}

	log.Printf("imx6_caam: ecdsap256 matches crypto/ecdsa")

	priv, _ = ecdsa.GenerateKey(ecc.P256k1(), rand.Reader)
	r, s, err = imx6ul.CAAM.Sign(priv, hash, nil)

	if err != nil {
		log.Printf("%v", err)
		return err
	}

	if !ecdsa.Verify(&priv.PublicKey, hash, r, s) {
		return fmt.Errorf("invalid secp256k1 signature")
	}

	log.Printf("imx6_caam: secp256k1 matches crypto/ecdsa")

	return
}

func testKeyDerivationCAAM(log *log.Logger) (err error) {
	key := make([]byte, sha256.Size)

	if err = imx6ul.CAAM.DeriveKey([]byte(testDiversifier), key); err != nil {
		return
	}

	if bytes.Compare(key, make([]byte, len(key))) == 0 {
		return fmt.Errorf("derivedKey all zeros")
	}

	// if the SoC is secure booted we can only print the result
	if imx6ul.SNVS.Available() {
		log.Printf("imx6_caam: OTPMK derived key %x", key)
		return
	}

	if strings.Compare(string(key), testVectorCAAM) != 0 {
		return fmt.Errorf("derivedKey:%x != testVector:%x", key, testVectorCAAM)
	}

	log.Printf("imx6_caam: derived test key %x", key)

	return
}

func caamTest() (tag string, res string) {
	tag = "imx6_caam"

	b := &strings.Builder{}
	log := log.New(b, "", 0)

	if !(imx6ul.Native && imx6ul.CAAM != nil) {
		log.Printf("skipping imx6_caam tests under emulation or unsupported hardware")
		return tag, b.String()
	}

	if err := testHashCAAM(log); err != nil {
		log.Printf("imx6_caam: hash error, %v", err)
	}

	for _, n := range []int{128, 192, 256} {
		if err := testCipherCAAM(n, log); err != nil {
			log.Printf("imx6_caam: cipher error, %v", err)
		}
	}

	if err := testKeyDerivationCAAM(log); err != nil {
		log.Printf("imx6_caam: key derivation error, %v", err)
	}

	if err := testSignatureCAAM(log); err != nil {
		log.Printf("imx6_caam: signature error, %v", err)
	}

	return tag, b.String()
}
