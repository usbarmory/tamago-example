// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"regexp"
	"runtime"

	"golang.org/x/term"

	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

const keySlot = 0

// NIST SP 800-38A test vectors
var (
	testVectorInput = "\x6b\xc1\xbe\xe2\x2e\x40\x9f\x96\xe9\x3d\x7e\x11\x73\x93\x17\x2a"
	testVectorIV    = "\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f"
	testVectorKey   = map[int]string{
		128: "\x2b\x7e\x15\x16\x28\xae\xd2\xa6\xab\xf7\x15\x88\x09\xcf\x4f\x3c",
		192: "\x8e\x73\xb0\xf7\xda\x0e\x64\x52\xc8\x10\xf3\x2b\x80\x90\x79\xe5\x62\xf8\xea\xd2\x52\x2c\x6b\x7b",
		256: "\x60\x3d\xeb\x10\x15\xca\x71\xbe\x2b\x73\xae\xf0\x85\x7d\x77\x81\x1f\x35\x2c\x07\x3b\x61\x08\xd7\x2d\x98\x10\xa3\x09\x14\xdf\xf4",
	}
	testVectorCipher = map[int]string{
		128: "\x76\x49\xab\xac\x81\x19\xb2\x46\xce\xe9\x8e\x9b\x12\xe9\x19\x7d",
		192: "\x4f\x02\x1d\xb2\x43\xbc\x63\x3d\x71\x78\x18\x3a\x9f\xa0\x71\xe8",
		256: "\xf5\x8c\x4c\x04\xd6\xe5\xf1\xba\x77\x9e\xab\xfb\x5f\x7b\xfb\xd6",
	}
	testVectorMAC = map[int]string{
		128: "\x07\x0a\x16\xb4\x6b\x4d\x41\x44\xf7\x9b\xdd\x9d\xd0\x4a\x28\x7c",
		192: "\x9e\x99\xa7\xbf\x31\xe7\x10\x90\x06\x62\xf6\x5e\x61\x7c\x51\x84",
		256: "\x28\xa7\x02\x3f\x45\x2e\x8f\x82\xbd\x4b\xf2\x8d\x8c\x37\xc3\x5c",
	}
)

func init() {
	Add(Cmd{
		Name:    "aes",
		Args:    3,
		Pattern: regexp.MustCompile(`^aes (\d+) (\d+)( soft)?$`),
		Syntax:  "<size> <sec> (soft)?",
		Help:    "benchmark CAAM/DCP hardware encryption",
		Fn:      aesCmd,
	})
}

func aesCmd(_ *Interface, _ *term.Terminal, arg []string) (res string, err error) {
	var fn func([]byte) (string, error)

	key := make([]byte, aes.BlockSize)
	iv := make([]byte, aes.BlockSize)

	switch {
	case len(arg[2]) > 0:
		block, err := aes.NewCipher(key)

		if err != nil {
			return "", err
		}

		fn = func(buf []byte) (_ string, err error) {
			cbc := cipher.NewCBCEncrypter(block, iv)
			cbc.CryptBlocks(buf, buf)
			runtime.Gosched()
			return
		}
	case imx6ul.CAAM != nil:
		fn = func(buf []byte) (_ string, err error) {
			err = imx6ul.CAAM.Encrypt(buf, key, iv)
			return
		}
	case imx6ul.DCP != nil:
		fn = func(buf []byte) (_ string, err error) {
			_ = imx6ul.DCP.SetKey(keySlot, key)
			err = imx6ul.DCP.Encrypt(buf, keySlot, iv)
			return
		}
	default:
		err = fmt.Errorf("unsupported hardware, use `aes %s %s soft` to disable hardware acceleration", arg[0], arg[1])
		return
	}

	return cipherCmd(arg, "aes-128 cbc", fn)
}
