// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Adapted from go/src/crypto/ecdsa/ecdsa_test.go

package cmd

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"log"
	"time"
)

func testSignAndVerify(c elliptic.Curve, tag string) {
	start := time.Now()
	log.Printf("ECDSA sign and verify with p%d ... ", c.Params().BitSize)

	priv, _ := ecdsa.GenerateKey(c, rand.Reader)

	hashed := []byte("testing")
	r, s, err := ecdsa.Sign(rand.Reader, priv, hashed)
	if err != nil {
		log.Printf("%s: error signing: %s", tag, err)
		return
	}

	if !ecdsa.Verify(&priv.PublicKey, hashed, r, s) {
		log.Printf("%s: Verify failed", tag)
	}

	hashed[0] ^= 0xff
	if ecdsa.Verify(&priv.PublicKey, hashed, r, s) {
		log.Printf("%s: Verify always works!", tag)
	}

	log.Printf("ECDSA sign and verify with p%d took %s", c.Params().BitSize, time.Since(start))
}

func ecdsaTest() {
	msg("ecdsa")
	testSignAndVerify(elliptic.P224(), "p224")
	testSignAndVerify(elliptic.P256(), "p256")
}
