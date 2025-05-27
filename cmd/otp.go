// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build mx6ullevk || usbarmory

package cmd

import (
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/usbarmory/crucible/fusemap"
	"github.com/usbarmory/crucible/otp"

	"github.com/usbarmory/tamago-example/internal/hab"
	"github.com/usbarmory/tamago-example/shell"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
	"github.com/usbarmory/tamago/soc/nxp/snvs"
)

const habWarning = `
████████████████████████████████████████████████████████████████████████████████

                 !!!           READ CAREFULLY           !!!

This command activates secure boot (HAB) on your device SoC by permanent OTP
fusing of the argument SRK hash.

Fusing OTP's is an **irreversible** action that permanently fuses values on the
device. This means that your device will be able to only execute firmware
signed with the corresponding private keys after programming is completed.

In other words your device will stop acting as a generic purpose device and
will be converted to *exclusive use of your own signed firmware releases*.

████████████████████████████████████████████████████████████████████████████████
`

func init() {
	shell.Add(shell.Cmd{
		Name:    "hab",
		Args:    1,
		Pattern: regexp.MustCompile(`^hab ([[:xdigit:]]+)$`),
		Syntax:  "<srk table hash>",
		Help:    "HAB activation (use with extreme caution)",
		Fn:      habCmd,
	})

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

func habCmd(console *shell.Interface, arg []string) (res string, err error) {
	srk, err := hex.DecodeString(arg[0])

	if err != nil {
		return "", fmt.Errorf("invalid SRK hash, %v", err)
	}

	fmt.Fprintf(console.Output, habWarning)

	if !console.Confirm("Are you sure? (y/n) ") {
		return "command cancelled", nil
	}

	if !imx6ul.Native {
		return "", errors.New("unavailable under emulation")
	}

	ssm := imx6ul.SNVS.Monitor()

	if ssm.State != snvs.SSM_STATE_NONSECURE {
		return "", fmt.Errorf("invalid state (%#.4b)", ssm.State)
	}

	return "", hab.Activate(srk)
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
