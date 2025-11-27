// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build imx8mpevk || mx6ullevk || usbarmory

package cmd

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/usbarmory/crucible/fusemap"
	"github.com/usbarmory/crucible/otp"

	"github.com/usbarmory/tamago-example/shell"
)

var fuseMap *fusemap.FuseMap

func init() {
	shell.Add(shell.Cmd{
		Name:    "otp",
		Args:    2,
		Pattern: regexp.MustCompile(`^otp (\d+) (\d+)$`),
		Syntax:  "<bank> <word>",
		Help:    "OTP fuses display",
		Fn:      otpCmd,
	})
}

func readOTP(bank int, word int) (res string, err error) {
	var reg *fusemap.Register
	var val []byte

	if err = loadFuseMap(); err != nil {
		return
	}

	if OCOTP != nil {
		if val, err = otp.ReadOCOTP(OCOTP, bank, word, 0, 32); err != nil {
			return
		}
	}

	for _, reg = range fuseMap.Registers {
		if reg.Bank == bank && reg.Word == word {
			res = fmt.Sprintf("OTP bank:%d word:%d val:%#x\n\n", bank, word, val)
			res += reg.BitMap(val)
			return
		}
	}

	return "", errors.New("invalid OTP register")
}

func otpCmd(_ *shell.Interface, arg []string) (res string, err error) {
	bank, err := strconv.Atoi(arg[0])

	if err != nil {
		return "", fmt.Errorf("invalid bank, %v", err)
	}

	word, err := strconv.Atoi(arg[1])

	if err != nil {
		return "", fmt.Errorf("invalid word, %v", err)
	}

	return readOTP(bank, word)
}
