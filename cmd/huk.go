// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build mx6ullevk || usbarmory

package cmd

import (
	"crypto/aes"
	"crypto/sha256"
	"errors"
	"fmt"

	"golang.org/x/term"

	"github.com/usbarmory/tamago-example/shell"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

func init() {
	shell.Add(shell.Cmd{
		Name: "huk",
		Help: "CAAM/DCP hardware unique key derivation",
		Fn:   hukCmd,
	})
}

func hukCmd(_ *shell.Interface, _ *term.Terminal, arg []string) (res string, err error) {
	var key []byte

	if !imx6ul.Native {
		return "", errors.New("unsupported under emulation")
	}

	switch {
	case imx6ul.CAAM != nil:
		key = make([]byte, sha256.Size)
		err = imx6ul.CAAM.DeriveKey([]byte(testDiversifier), key)
		res = "CAAM DeriveKey"
	case imx6ul.DCP != nil:
		iv := make([]byte, aes.BlockSize)
		key, err = imx6ul.DCP.DeriveKey([]byte(testDiversifier), iv, -1)
		res = "DCP DeriveKey"
	default:
		err = errors.New("unsupported hardware")
	}

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s: %x", res, key), nil
}
