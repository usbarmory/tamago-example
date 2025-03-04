// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build amd64

package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"regexp"
	"runtime"
	"strconv"

	"github.com/usbarmory/tamago-example/shell"
	"github.com/usbarmory/tamago/amd64"
	"github.com/usbarmory/virtio-net"
)

var NIC *vnet.Net

func init() {
	shell.Add(shell.Cmd{
		Name:    "cpuid",
		Args:    2,
		Pattern: regexp.MustCompile(`^cpuid ([[:xdigit:]]+) ([[:xdigit:]]+)$`),
		Syntax:  "<leaf> <subleaf>",
		Help:    "display CPU capabilities",
		Fn:      cpuidCmd,
	})
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
	fmt.Fprintf(&res, "CPU ..........: %s\n", name)
	fmt.Fprintf(&res, "Frequency ....: %v GHz\n", float32(freq)/1e9)

	if NIC != nil {
		mac := NIC.Config().MAC
		fmt.Fprintf(&res, "VirtIO Net%d ..: %s\n", NIC.Index, net.HardwareAddr(mac[:]))
	}

	return res.String(), nil
}

func cpuidCmd(_ *shell.Interface, arg []string) (string, error) {
	var res bytes.Buffer

	leaf, err := strconv.ParseUint(arg[0], 16, 32)

	if err != nil {
		return "", fmt.Errorf("invalid leaf, %v", err)
	}

	subleaf, err := strconv.ParseUint(arg[1], 10, 32)

	if err != nil {
		return "", fmt.Errorf("invalid subleaf, %v", err)
	}

	cpu := amd64.CPU{}
	eax, ebx, ecx, edx := cpu.CPUID(uint32(leaf), uint32(subleaf))

	fmt.Fprintf(&res, "EAX      EBX      ECX      EDX\n")
	fmt.Fprintf(&res, "%08x %08x %08x %08x\n", eax, ebx, ecx, edx)

	return res.String(), nil
}

func rebootCmd(_ *shell.Interface, _ []string) (_ string, err error) {
	return "", errors.New("unimplemented")
}

func cryptoTest() {
	spawn(btcdTest)
	spawn(kemTest)
}

func storageTest() {
	return
}

func HasNetwork() (usb bool, eth bool) {
	return false, true
}
