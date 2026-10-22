// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build gcp

package cmd

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"

	"github.com/usbarmory/tamago/board/google/gcp"
	"github.com/usbarmory/tamago/soc/intel/nvme"
	"github.com/usbarmory/tamago/soc/intel/pci"

	"github.com/usbarmory/tamago-example/network"
	"github.com/usbarmory/tamago-example/shell"
)

const nvmeVector = network.VIRTIO_NET0_IRQ + 1

func init() {
	shell.Add(shell.Cmd{
		Name:    "nvme",
		Args:    4,
		Pattern: regexp.MustCompile(`^nvme (\d) (\d) ([[:xdigit:]]+) (\d+)$`),
		Syntax:  "<bus> <n> <hex addr> <size>",
		Help:    "NVMe read",
		Fn:      nvmeCmd,
	})
}

func nvmeDetect(bus int, n int) (drive *nvme.NVMe) {
	dev := pci.ProbeFunction(bus, n, nvme.PCI_CLASS, nvme.PCI_SUBCLASS, nvme.PCI_PROG_IF)

	if dev == nil {
		return
	}

	return &nvme.NVMe{
		Device:     dev,
		Queues:     4,
		QueueDepth: 128,
		Vector:     nvmeVector,
		CPU:        gcp.AMD64,
	}
}

func nvmeRead(drive *nvme.NVMe, addr int64, size int) (buf []byte, err error) {
	isr := gcp.AMD64.ISR
	defer func() { gcp.AMD64.ISR = isr }()

	gcp.AMD64.ISR = func(irq int) {
		if !drive.ServiceInterrupt(irq) {
			isr(irq)
		}
	}

	if err = drive.Init(); err != nil {
		return nil, fmt.Errorf("could not initialize nvme%d %v", drive.Index, err)
	}
	defer drive.Shutdown()

	ns := drive.Namespaces[0]
	bs := int(ns.BlockSize)
	s := int(size)

	// align read to block size
	if r := s % bs; r > 0 {
		s += bs - r
	}

	buf = make([]byte, s)
	n, err := ns.ReadAt(buf, int64(addr))

	return buf[:min(size, n)], err
}

func nvmeCmd(_ *shell.Interface, arg []string) (res string, err error) {
	bus, err := strconv.ParseUint(arg[0], 10, 8)

	if err != nil {
		return "", fmt.Errorf("invalid bus index: %v", err)
	}

	n, err := strconv.ParseUint(arg[1], 10, 8)

	if err != nil {
		return "", fmt.Errorf("invalid device index: %v", err)
	}

	addr, err := strconv.ParseUint(arg[2], 16, 32)

	if err != nil {
		return "", fmt.Errorf("invalid address: %v", err)
	}

	size, err := strconv.ParseUint(arg[3], 10, 32)

	if err != nil {
		return "", fmt.Errorf("invalid size: %v", err)
	}

	if size > maxBufferSize {
		return "", fmt.Errorf("size argument must be <= %d", maxBufferSize)
	}

	drive := nvmeDetect(int(bus), int(n))

	if drive == nil {
		return "", fmt.Errorf("nvme%d not found", n)
	}

	buf, err := nvmeRead(drive, int64(addr), int(size))
	res = fmt.Sprintf("%s", hex.Dump(buf))

	return
}
