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
	"fmt"
	"log"
	"regexp"
	"time"

	"golang.org/x/term"

	"github.com/usbarmory/tamago/soc/nxp/caam"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

func init() {
	Add(Cmd{
		Name:    "ecdsa",
		Args:    2,
		Pattern: regexp.MustCompile(`^ecdsa (\d+)( soft)?$`),
		Syntax:  "<sec> (soft)?",
		Help:    "benchmark CAAM/DCP hardware signing",
		Fn:      ecdsaCmd,
	})
}

func ecdsaCmd(_ *Interface, _ *term.Terminal, arg []string) (res string, err error) {
	var fn func([]byte) (string, error)

	curve := elliptic.P256()
	priv, _ := ecdsa.GenerateKey(curve, rand.Reader)

	arg = append([]string{fmt.Sprintf("%d", curve.Params().BitSize/8)}, arg...)

	switch {
	case len(arg[2]) > 0:
		fn = func(buf []byte) (_ string, err error) {
			_, _, err = ecdsa.Sign(rand.Reader, priv, buf)
			return
		}
	case imx6ul.CAAM != nil:
		pdb := &caam.SignPDB{}
		defer pdb.Free()

		if err = pdb.Init(priv); err != nil {
			return
		}

		fn = func(buf []byte) (_ string, err error) {
			_, _, err = imx6ul.CAAM.Sign(nil, buf, pdb)
			return
		}
	default:
		err = fmt.Errorf("unsupported hardware, use `ecdsa %s soft` to disable hardware acceleration", arg[1])
		return
	}

	return cipherCmd(arg, "ecdsap256", fn)
}

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
