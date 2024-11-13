// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build mx6ullevk || usbarmory

package cmd

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"regexp"

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
