// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"crypto/aes"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"golang.org/x/term"

	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

const testDiversifier = "\xde\xad\xbe\xef"

func init() {
	Add(Cmd{
		Name: "huk",
		Help: "CAAM/DCP hardware unique key derivation",
		Fn:   hukCmd,
	})
}

func hukCmd(_ *term.Terminal, arg []string) (res string, err error) {
	if !imx6ul.Native {
		return "", errors.New("unsupported under emulation")
	}

	var key []byte

	switch {
	case imx6ul.CAAM != nil:
		key, err = imx6ul.CAAM.MasterKeyVerification()
		res = "CAAM MKV BKEK"
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

func cipherCmd(arg []string, tag string, fn func(buf []byte) (string, error)) (res string, err error) {
	size, err := strconv.Atoi(arg[0])

	if err != nil {
		return "", fmt.Errorf("invalid size, %v", err)
	}

	sec, err := strconv.Atoi(arg[1])

	if err != nil {
		return "", fmt.Errorf("invalid duration, %v", err)
	}

	log.Printf("Doing %s for %ds on %d size blocks", tag, sec, size)

	n := 0
	buf := make([]byte, size)
	start := time.Now()

	for run, timeout := true, time.After(time.Duration(sec)*time.Second); run; {
		if res, err = fn(buf); err != nil {
			return
		}

		n++

		select {
		case <-timeout:
			run = false
		default:
		}
	}

	return fmt.Sprintf("%d %s's in %s", n, tag, time.Since(start)), nil
}
