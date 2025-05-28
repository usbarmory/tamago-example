// Copyright 2022 The Armored Witness OS authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hab

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"

	"github.com/usbarmory/tamago/soc/nxp/imx6ul"

	"github.com/usbarmory/crucible/otp"
	"github.com/usbarmory/crucible/util"
)

// Activate enables secure boot by following the procedure described at:
//
//	https://github.com/usbarmory/usbarmory/wiki/Secure-boot-(Mk-II)#activating-hab
//
// IMPORTANT: enabling secure boot functionality on NXP i.MX SoCs, unlike
// similar features on modern PCs, is an irreversible action that permanently
// fuses verification keys hashes on the device. This means that any errors in
// the process or loss of the signing PKI will result in a bricked device
// incapable of executing unsigned code. This is a security feature, not a bug.
func Activate(srk []byte) (err error) {
	switch {
	case imx6ul.SNVS.Available():
		return errors.New("HAB already enabled")
	case len(srk) != sha256.Size:
		return errors.New("invalid SRK")
	case bytes.Equal(srk, make([]byte, sha256.Size)):
		return errors.New("invalid SRK")
	default:
		// Enable High Assurance Boot (i.e. secure boot)
		return hab(srk)
	}

	return
}

func fuse(name string, bank int, word int, off int, size int, val []byte) error {
	log.Printf("fusing %s bank:%d word:%d off:%d size:%d val:%x", name, bank, word, off, size, val)

	if res, err := otp.ReadOCOTP(bank, word, off, size); err != nil {
		return fmt.Errorf("read error for %s, res:%x err:%v\n", name, res, err)
	} else {
		log.Printf("  pre-read val: %x", res)
	}

	if err := otp.BlowOCOTP(bank, word, off, size, val); err != nil {
		return err
	}

	if res, err := otp.ReadOCOTP(bank, word, off, size); err != nil || !bytes.Equal(val, res) {
		return fmt.Errorf("readback error for %s, val:%x res:%x err:%v\n", name, val, res, err)
	}

	return nil
}

func hab(srk []byte) (err error) {
	if len(srk) != sha256.Size {
		return errors.New("fatal error, invalid SRK hash")
	}

	// fuse HAB public keys hash
	if err = fuse("SRK_HASH", 3, 0, 0, 256, util.SwitchEndianness(srk)); err != nil {
		return
	}

	// lock HAB public keys hash
	if err = fuse("SRK_LOCK", 0, 0, 14, 1, []byte{1}); err != nil {
		return
	}

	// set device in Closed Configuration (IMX6ULRM Table 8-2, p245)
	if err = fuse("SEC_CONFIG", 0, 6, 0, 2, []byte{0b11}); err != nil {
		return
	}

	// disable NXP reserved mode (IMX6ULRM 8.2.6, p244)
	if err = fuse("DIR_BT_DIS", 0, 6, 3, 1, []byte{1}); err != nil {
		return
	}

	// Disable debugging features (IMX6ULRM Table 5-9, p216)

	// disable Secure JTAG controller
	if err = fuse("SJC_DISABLE", 0, 6, 20, 1, []byte{1}); err != nil {
		return
	}

	// disable JTAG debug mode
	if err = fuse("JTAG_SMODE", 0, 6, 22, 2, []byte{0b11}); err != nil {
		return
	}

	// disable HAB ability to enable JTAG
	if err = fuse("JTAG_HEO", 0, 6, 27, 1, []byte{1}); err != nil {
		return
	}

	// disable tracing
	if err = fuse("KTE", 0, 6, 26, 1, []byte{1}); err != nil {
		return
	}

	// Further reduce the attack surface

	// disable Serial Download Protocol (SDP) READ_REGISTER command (IMX6ULRM 8.9.3, p310)
	if err = fuse("SDP_READ_DISABLE", 0, 6, 18, 1, []byte{1}); err != nil {
		return
	}

	// disable SDP over UART (IMX6ULRM 8.9, p305)
	if err = fuse("UART_SERIAL_DOWNLOAD_DISABLE", 0, 7, 4, 1, []byte{1}); err != nil {
		return
	}

	return
}
