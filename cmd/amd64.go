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
	"strings"
	"sync"

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

	shell.Add(shell.Cmd{
		Name:    "smp",
		Args:    1,
		Pattern: regexp.MustCompile(`^smp (\d+)$`),
		Syntax:  "<n>",

		Help:    "launch SMP test",
		Fn:      smpCmd,
	})
}

func mem(start uint, size int, w []byte) (b []byte) {
	return memCopy(start, size, w)
}

func infoCmd(_ *shell.Interface, _ []string) (string, error) {
	var res bytes.Buffer

	ramStart, ramEnd := runtime.MemRegion()
	name, freq := Target()

	fmt.Fprintf(&res, "Runtime ......: %s %s/%s GOMAXPROCS=%d\n", runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.GOMAXPROCS(-1))
	fmt.Fprintf(&res, "RAM ..........: %#08x-%#08x (%d MiB)\n", ramStart, ramEnd, (ramEnd-ramStart)/(1024*1024))
	fmt.Fprintf(&res, "Board ........: %s\n", boardName)
	fmt.Fprintf(&res, "CPU ..........: %s\n", name)
	fmt.Fprintf(&res, "Cores ........: %d\n", amd64.NumCPU())
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

func smpCmd(console *shell.Interface, arg []string) (string, error) {
	var res bytes.Buffer
	var wg sync.WaitGroup
	var cc sync.Map

	n, err := strconv.Atoi(arg[0])

	if err != nil {
		return "", fmt.Errorf("invalid goroutine count: %v", err)
	}

	ncpu := runtime.NumCPU()
	wg.Add(n)

	if runtime.ProcID == nil || runtime.Task == nil {
		return "", errors.New("no SMP detected")
	}

	fmt.Fprintf(console.Output,"%d cores detected, launching %d goroutines from CPU%2d\n", ncpu, n, runtime.ProcID())

	for i := 0; i < n; i++ {
		go func() {
			cpu := runtime.ProcID()

			for {
				if actual, loaded := cc.LoadOrStore(cpu, 1); loaded {
					if cc.CompareAndSwap(cpu, actual.(int), actual.(int)+1) {
						break
					}
				} else {
					break
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()

	var total int

	cc.Range(func(cpu any, count any) bool {
		total += count.(int)
		fmt.Fprintf(&res, "CPU%2d %3d:%s\n", cpu.(uint64), count.(int), strings.Repeat("░", count.(int)))
		return true
	})

	fmt.Fprintf(&res, "Total %3d\n", total)

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
