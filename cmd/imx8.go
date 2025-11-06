// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build imx8mpevk

package cmd

import (
	"bytes"
	_ "embed"
	"fmt"
	"runtime"
	_ "unsafe"

	"github.com/usbarmory/crucible/fusemap"

	"github.com/usbarmory/tamago/soc/nxp/dcp"
	"github.com/usbarmory/tamago/soc/nxp/imx8mp"

	"github.com/usbarmory/tamago-example/internal/semihosting"
	"github.com/usbarmory/tamago-example/shell"
)

//go:linkname ramSize runtime.ramSize
var ramSize uint = 0x20000000 // 512MB

var (
	// stub
	DCP *dcp.DCP

	CAAM  = imx8mp.CAAM
	SNVS  = imx8mp.SNVS
	OCOTP = imx8mp.OCOTP

	//go:embed IMX8MP.yaml
	IMX8MPFusemapYAML []byte
)

func loadFuseMap() (err error) {
	if fuseMap != nil {
		return
	}

	switch imx8mp.Model() {
	case "i.MX8MP":
		fuseMap, err = fusemap.Parse(IMX8MPFusemapYAML)
	}

	return
}

func init() {
	runtime.Exit = func(_ int32) {
		semihosting.Exit()
	}
}

func date(epoch int64) {
	imx8mp.ARM64.SetTime(epoch)
}

func uptime() (ns int64) {
	return imx8mp.ARM64.GetTime() - imx8mp.ARM64.TimerOffset
}

func mem(start uint, size int, w []byte) (b []byte) {
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

	return res.String(), nil
}

func cryptoTest() {
	spawn(btcdTest)
	spawn(kemTest)
	spawn(caamTest)

	return
}

func Target() (name string, freq uint32) {
	name = imx8mp.Model()

	if !imx8mp.Native {
		name += " (emulated)"
	}

	freq = imx8mp.ARMFreq()

	return
}
