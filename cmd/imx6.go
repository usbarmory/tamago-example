// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build mx6ullevk || usbarmory

package cmd

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	_ "unsafe"

	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
	"github.com/usbarmory/tamago/soc/nxp/snvs"

	"github.com/usbarmory/tamago-example/internal/semihosting"
	"github.com/usbarmory/tamago-example/shell"
)

const (
	// Override standard memory allocation as having concurrent USB and
	// Ethernet interfaces requires more than what the iRAM can handle.
	dmaSize  = 0xa00000 // 10MB
	dmaStart = 0xa0000000 - dmaSize

	romStart = 0x00000000
	romSize  = 0x17000
)

//go:linkname ramSize runtime.ramSize
var ramSize uint = 0x20000000 - dmaSize // 512MB - 10MB

var (
	DCP  = imx6ul.DCP
	CAAM = imx6ul.CAAM
	SNVS = imx6ul.SNVS
)

func init() {
	dma.Init(dmaStart, dmaSize)

	if !imx6ul.Native {
		runtime.Exit = func(_ int32) {
			semihosting.Exit()
		}

		return
	}

	switch imx6ul.Model() {
	case "i.MX6ULL", "i.MX6ULZ":
		imx6ul.SetARMFreq(imx6ul.FreqMax)
	case "i.MX6UL":
		imx6ul.SetARMFreq(imx6ul.Freq528)
	}

	shell.Add(shell.Cmd{
		Name:    "freq",
		Args:    1,
		Pattern: regexp.MustCompile(`^freq (198|396|528|792|900)$`),
		Help:    "change ARM core frequency",
		Syntax:  "(198|396|528|792|900)",
		Fn:      freqCmd,
	})

	// This example policy sets the maximum delay between violation
	// detection and hard failure, on the i.MX6UL SNVS re-initialization
	// with invalid calibration data (e.g. SNVS.Init(0)) can be used to
	// test tamper detection.
	imx6ul.SNVS.SetPolicy(
		snvs.SecurityPolicy{
			Clock:             true,
			Temperature:       true,
			Voltage:           true,
			SecurityViolation: true,
			HardFail:          true,
			HAC:               0xffffffff,
		},
	)

	if imx6ul.CAAM != nil {
		imx6ul.CAAM.DeriveKeyMemory, _ = dma.NewRegion(imx6ul.OCRAM_START, imx6ul.OCRAM_SIZE, false)
	}
}

func date(epoch int64) {
	imx6ul.ARM.SetTime(epoch)
}

func uptime() (ns int64) {
	return imx6ul.ARM.GetTime() - imx6ul.ARM.TimerOffset
}

func mem(start uint, size int, w []byte) (b []byte) {
	// temporarily map page zero if required
	if z := uint32(1 << 20); uint32(start) < z {
		imx6ul.ARM.ConfigureMMU(0, z, 0, (arm.TTE_AP_001<<10)|arm.TTE_SECTION)
		defer imx6ul.ARM.ConfigureMMU(0, z, 0, 0)
	}

	return memCopy(start, size, w)
}

func infoCmd(_ *shell.Interface, _ []string) (string, error) {
	var res bytes.Buffer

	ramStart, ramEnd := runtime.MemRegion()
	name, freq := Target()

	fmt.Fprintf(&res, "Runtime ......: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(&res, "RAM ..........: %#08x-%#08x (%d MiB)\n", ramStart, ramEnd, (ramEnd-ramStart)/(1024*1024))
	fmt.Fprintf(&res, "Board ........: %s\n", boardName)
	fmt.Fprintf(&res, "SoC ..........: %s\n", name)
	fmt.Fprintf(&res, "Frequency ....: %v MHz\n", float32(freq)/1e6)

	if NIC != nil {
		fmt.Fprintf(&res, "ENET%d ........: %s %d\n", NIC.Index, NIC.MAC, NIC.Stats)
	}

	if !imx6ul.Native {
		return res.String(), nil
	}

	ssm := imx6ul.SNVS.Monitor()
	fmt.Fprintf(&res,
		"SSM ..........: state:%#.4b clk:%v tmp:%v vcc:%v hac:%d\n",
		ssm.State, ssm.Clock, ssm.Temperature, ssm.Voltage, ssm.HAC,
	)

	if imx6ul.CAAM != nil {
		cs, err := imx6ul.CAAM.RSTA()
		fmt.Fprintf(&res, "RTIC .........: state:%#.4b err:%v\n", cs, err)
	}

	rom := mem(romStart, romSize, nil)
	fmt.Fprintf(&res, "Boot ROM hash : %x\n", sha256.Sum256(rom))
	fmt.Fprintf(&res, "Secure boot ..: %v\n", imx6ul.SNVS.Available())

	fmt.Fprintf(&res, "Unique ID ....: %X\n", imx6ul.UniqueID())
	fmt.Fprintf(&res, "SDP ..........: %v\n", imx6ul.SDP)
	fmt.Fprintf(&res, "Temperature ..: %f\n", imx6ul.TEMPMON.Read())

	return res.String(), nil
}

func freqCmd(_ *shell.Interface, arg []string) (res string, err error) {
	var mhz uint64

	if mhz, err = strconv.ParseUint(arg[0], 10, 32); err != nil {
		return
	}

	err = imx6ul.SetARMFreq(uint32(mhz))
	_, freq := Target()

	return fmt.Sprintf("%v MHz", float32(freq)/1e6), err
}

func cryptoTest() {
	spawn(btcdTest)
	spawn(kemTest)
	spawn(caamTest)
	spawn(dcpTest)

	return
}

func Target() (name string, freq uint32) {
	name = imx6ul.Model()

	if !imx6ul.Native {
		name += " (emulated)"
	}

	freq = imx6ul.ARMFreq()

	return
}
