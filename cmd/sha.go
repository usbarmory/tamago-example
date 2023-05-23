// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build mx6ullevk || usbarmory
// +build mx6ullevk usbarmory

package cmd

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"regexp"
	"runtime"

	"golang.org/x/term"

	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

const testVectorNIST3 = "\xcd\xc7\x6e\x5c\x99\x14\xfb\x92\x81\xa1\xc7\xe2\x84\xd7\x3e\x67\xf1\x80\x9a\x48\xa4\x97\x20\x0e\x04\x6d\x39\xcc\xc7\x11\x2c\xd0"

func init() {
	Add(Cmd{
		Name:    "sha",
		Args:    2,
		Pattern: regexp.MustCompile(`^sha (\d+) (\d+)$`),
		Syntax:  "<size> <sec>",
		Help:    "benchmark CAAM/DCP hardware hashing",
		Fn:      shaCmd,
	})
}

func shaCmd(_ *term.Terminal, arg []string) (res string, err error) {
	fn := func(buf []byte) (res string, err error) {
		var sum [32]byte

		switch {
		case !imx6ul.Native:
			sum = sha256.Sum256(buf)
			runtime.Gosched()
		case imx6ul.CAAM != nil:
			sum, err = imx6ul.CAAM.Sum256(buf)
		case imx6ul.DCP != nil:
			sum, err = imx6ul.DCP.Sum256(buf)
		default:
			err = errors.New("unsupported hardware")
		}

		return fmt.Sprintf("%x", sum), err
	}

	return cipherCmd(arg, "sha256", fn)
}
