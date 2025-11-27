// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build imx8mpevk || mx6ullevk || usbarmory

package cmd

import (
	"crypto/aes"
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/usbarmory/tamago-example/shell"
)

func init() {
	shell.Add(shell.Cmd{
		Name: "huk",
		Help: "CAAM/DCP hardware unique key derivation",
		Fn:   hukCmd,
	})
}

func hukCmd(_ *shell.Interface, arg []string) (res string, err error) {
	var key []byte

	switch {
	case CAAM != nil:
		key = make([]byte, sha256.Size)
		err = CAAM.DeriveKey([]byte(testDiversifier), key)
		res = "CAAM DeriveKey"
	case DCP != nil:
		iv := make([]byte, aes.BlockSize)
		key, err = DCP.DeriveKey([]byte(testDiversifier), iv, -1)
		res = "DCP DeriveKey"
	default:
		err = errors.New("unavailable")
	}

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s: %x", res, key), nil
}
