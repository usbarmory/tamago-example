// https://github.com/f-secure-foundry/tamago-example
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	_ "embed"
	"errors"
	"fmt"

	"github.com/f-secure-foundry/crucible/fusemap"
	"github.com/f-secure-foundry/crucible/otp"
)

//go:embed IMX6ULL.yaml
var IMX6ULLFusemapYAML []byte

var IMX6ULLFusemap *fusemap.FuseMap

func loadFuseMap() (err error) {
	if IMX6ULLFusemap != nil {
		return
	}

	if len(IMX6ULLFusemapYAML) == 0 {
		return errors.New("fusemap not available")
	}

	IMX6ULLFusemap, err = fusemap.Parse(IMX6ULLFusemapYAML)

	return
}

func readOTP(bank int, word int) (res string, err error) {
	var reg *fusemap.Register

	if err := loadFuseMap(); err != nil {
		return "", err
	}

	val, err := otp.ReadOCOTP(bank, word, 0, 32)

	if err != nil {
		return "", err
	}

	for _, reg = range IMX6ULLFusemap.Registers {
		if reg.Bank == bank && reg.Word == word {
			res = fmt.Sprintf("OTP bank:%d word:%d val:%#x\n\n", bank, word, val)
			res += reg.BitMap(val)
			return
		}
	}

	return "", errors.New("invalid OTP register")
}
