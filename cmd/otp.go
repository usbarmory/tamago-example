// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build mx6ullevk || usbarmory

package cmd

import (
	_ "embed"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"golang.org/x/term"

	"github.com/usbarmory/crucible/fusemap"
	"github.com/usbarmory/crucible/otp"

	"github.com/usbarmory/tamago-example/shell"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

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

var (
	//go:embed IMX6UL.yaml
	IMX6ULFusemapYAML []byte
	//go:embed IMX6ULL.yaml
	IMX6ULLFusemapYAML []byte

	fuseMap *fusemap.FuseMap
)

func loadFuseMap() (err error) {
	if fuseMap != nil {
		return
	}

	switch imx6ul.Model() {
	case "i.MX6ULL", "i.MX6ULZ":
		fuseMap, err = fusemap.Parse(IMX6ULLFusemapYAML)
	case "i.MX6UL":
		fuseMap, err = fusemap.Parse(IMX6ULFusemapYAML)
	}

	return
}

func readOTP(bank int, word int) (res string, err error) {
	var reg *fusemap.Register
	var val []byte

	if err = loadFuseMap(); err != nil {
		return
	}

	if imx6ul.Native {
		if val, err = otp.ReadOCOTP(bank, word, 0, 32); err != nil {
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

func otpCmd(_ *shell.Interface, _ *term.Terminal, arg []string) (res string, err error) {
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
