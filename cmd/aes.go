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

var testVector = map[int]string{
	128: "\x66\xe9\x4b\xd4\xef\x8a\x2c\x3b\x88\x4c\xfa\x59\xca\x34\x2b\x2e",
	192: "\xaa\xe0\x69\x92\xac\xbf\x52\xa3\xe8\xf4\xa9\x6e\xc9\x30\x0b\xd7",
	256: "\xdc\x95\xc0\x78\xa2\x40\x89\x89\xad\x48\xa2\x14\x92\x84\x20\x87",
}

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
			err = imx6ul.CAAM.Encrypt(buf, key, iv)
		case imx6ul.DCP != nil:
			_ = imx6ul.DCP.SetKey(keySlot, key)
			err = imx6ul.DCP.Encrypt(buf, keySlot, iv)
		default:
			err = errors.New("unsupported hardware")
		}

		return
	}

	return cipherCmd(arg, "aes-128 cbc", fn)
}
