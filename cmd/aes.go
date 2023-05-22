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
	"crypto/cipher"
	"errors"
	"regexp"
	"runtime"

	"golang.org/x/term"

	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

const keySlot = 0

func init() {
	Add(Cmd{
		Name:    "aes",
		Args:    2,
		Pattern: regexp.MustCompile(`^aes (\d+) (\d+)$`),
		Syntax:  "<size> <sec>",
		Help:    "benchmark CAAM/DCP hardware encryption",
		Fn:      aesCmd,
	})
}

func aesCmd(_ *term.Terminal, arg []string) (res string, err error) {
	key := make([]byte, aes.BlockSize)
	iv := make([]byte, aes.BlockSize)

	block, err := aes.NewCipher(key)

	if err != nil {
		return
	}

	fn := func(buf []byte) (_ string, err error) {
		switch {
		case !imx6ul.Native:
			cbc := cipher.NewCBCEncrypter(block, iv)
			cbc.CryptBlocks(buf, buf)
			runtime.Gosched()
		case imx6ul.CAAM != nil:
			err = errors.New("unsupported")
		case imx6ul.DCP != nil:
			_ = imx6ul.DCP.SetKey(keySlot, key)
			err = imx6ul.DCP.Decrypt(buf, keySlot, iv)
		default:
			err = errors.New("unsupported hardware")
		}

		return
	}

	return cipherCmd(arg, "aes-128 cbc", fn)
}
