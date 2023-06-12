// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"regexp"
	"runtime"

	"golang.org/x/term"

	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

// FIPS 180-2 test vector
var (
	testVectorSHAInput = bytes.Repeat([]byte("a"), 1000000)
	testVectorSHA      = "\xcd\xc7\x6e\x5c\x99\x14\xfb\x92\x81\xa1\xc7\xe2\x84\xd7\x3e\x67\xf1\x80\x9a\x48\xa4\x97\x20\x0e\x04\x6d\x39\xcc\xc7\x11\x2c\xd0"
)

func init() {
	Add(Cmd{
		Name:    "sha",
		Args:    3,
		Pattern: regexp.MustCompile(`^sha (\d+) (\d+)( soft)?$`),
		Syntax:  "<size> <sec> (soft)?",
		Help:    "benchmark CAAM/DCP hardware hashing",
		Fn:      shaCmd,
	})
}

func shaCmd(_ *Interface, _ *term.Terminal, arg []string) (res string, err error) {
	var fn func([]byte) (string, error)

	switch {
	case len(arg[2]) > 0:
		fn = func(buf []byte) (res string, err error) {
			sha256.Sum256(buf)
			runtime.Gosched()
			return
		}
	case imx6ul.CAAM != nil:
		fn = func(buf []byte) (res string, err error) {
			_, err = imx6ul.CAAM.Sum256(buf)
			return
		}
	case imx6ul.DCP != nil:
		fn = func(buf []byte) (res string, err error) {
			_, err = imx6ul.DCP.Sum256(buf)
			return
		}
	default:
		err = fmt.Errorf("unsupported hardware, use `sha %s %s soft` to disable hardware acceleration", arg[0], arg[1])
		return
	}

	return cipherCmd(arg, "sha256", fn)
}
