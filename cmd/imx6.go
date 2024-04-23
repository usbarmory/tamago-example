// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build mx6ullevk || usbarmory
// +build mx6ullevk usbarmory

package cmd

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"regexp"
	"runtime"
	"strconv"

	"golang.org/x/term"

	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
	"github.com/usbarmory/tamago/soc/nxp/snvs"
)

const (
	romStart = 0x00000000
	romSize  = 0x17000
)

func Target() (t string) {
	t = fmt.Sprintf("%s %v MHz", imx6ul.Model(), float32(imx6ul.ARMFreq())/1000000)

	if !imx6ul.Native {
		t += " (emulated)"
	}

	return
}

func init() {
	if !imx6ul.Native {
		return
	}

	switch imx6ul.Model() {
	case "i.MX6ULL", "i.MX6ULZ":
		imx6ul.SetARMFreq(imx6ul.FreqMax)
	case "i.MX6UL":
		imx6ul.SetARMFreq(imx6ul.Freq528)
	}

	Add(Cmd{
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
}

func date(epoch int64) {
	imx6ul.ARM.SetTimer(epoch)
}

func uptime() (ns int64) {
	return int64(imx6ul.ARM.TimerFn() * imx6ul.ARM.TimerMultiplier)
}

func mem(start uint, size int, w []byte) (b []byte) {
	// temporarily map page zero if required
	if z := uint32(1 << 20); uint32(start) < z {
		imx6ul.ARM.ConfigureMMU(0, z, 0, (arm.TTE_AP_001<<10)|arm.TTE_SECTION)
		defer imx6ul.ARM.ConfigureMMU(0, z, 0, 0)
	}

	return memCopy(start, size, w)
}

func infoCmd(_ *Interface, _ *term.Terminal, _ []string) (string, error) {
	var res bytes.Buffer

	ramStart, ramEnd := runtime.MemRegion()

	res.WriteString(fmt.Sprintf("Runtime ......: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH))
	res.WriteString(fmt.Sprintf("RAM ..........: %#08x-%#08x (%d MiB)\n", ramStart, ramEnd, (ramEnd-ramStart)/(1024*1024)))
	res.WriteString(fmt.Sprintf("Board ........: %s\n", boardName))
	res.WriteString(fmt.Sprintf("SoC ..........: %s\n", Target()))

	if NIC != nil {
		res.WriteString(fmt.Sprintf("ENET%d ........: %s %d\n", NIC.Index, NIC.MAC, NIC.Stats))
	}

	if !imx6ul.Native {
		return res.String(), nil
	}

	ssm := imx6ul.SNVS.Monitor()
	res.WriteString(fmt.Sprintf(
		"SSM Status ...: state:%#.4b clk:%v tmp:%v vcc:%v hac:%d\n",
		ssm.State, ssm.Clock, ssm.Temperature, ssm.Voltage, ssm.HAC,
	))

	if imx6ul.CAAM != nil {
		cs, err := imx6ul.CAAM.RSTA()
		res.WriteString(fmt.Sprintf("RTIC Status ..: cs:%#.2b err:%v\n", cs, err))
	}

	rom := mem(romStart, romSize, nil)
	res.WriteString(fmt.Sprintf("Boot ROM hash : %x\n", sha256.Sum256(rom)))
	res.WriteString(fmt.Sprintf("Secure boot ..: %v\n", imx6ul.SNVS.Available()))

	res.WriteString(fmt.Sprintf("Unique ID ....: %X\n", imx6ul.UniqueID()))
	res.WriteString(fmt.Sprintf("SDP ..........: %v\n", imx6ul.SDP))
	res.WriteString(fmt.Sprintf("Temperature ..: %f\n", imx6ul.TEMPMON.Read()))

	return res.String(), nil
}

func freqCmd(_ *Interface, _ *term.Terminal, arg []string) (res string, err error) {
	var mhz uint64

	if mhz, err = strconv.ParseUint(arg[0], 10, 32); err != nil {
		return
	}

	err = imx6ul.SetARMFreq(uint32(mhz))

	return Target(), err
}

func cryptoTest() {
	spawn(btcdTest)
	spawn(kemTest)
	spawn(caamTest)
	spawn(dcpTest)
}
