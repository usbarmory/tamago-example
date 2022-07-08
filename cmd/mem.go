// https://github.com/usbarmory/tamago-example
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	mathrand "math/rand"
	"regexp"
	"runtime"
	"strconv"

	"golang.org/x/term"

	"github.com/usbarmory/tamago/dma"
)

const (
	runs      = 9
	chunksMax = 50
	fillSize  = 160 * 1024 * 1024
)

func init() {
	Add(Cmd{
		Name: "md",
		Args: 2,
		Pattern: regexp.MustCompile(`^md ([[:xdigit:]]+) (\d+)`),
		Syntax: "<hex offset> <size>",
		Help: "memory display (use with caution)",
		Fn: memReadCmd,
	})

	Add(Cmd{
		Name: "mw",
		Args: 2,
		Pattern: regexp.MustCompile(`^mw ([[:xdigit:]]+) ([[:xdigit:]]+)`),
		Syntax: "<hex offset> <hex value>",
		Help: "memory write (use with caution)",
		Fn: memWriteCmd,
	})
}

func memCopy(start uint32, size int, w []byte) (b []byte) {
	mem := &dma.Region{
		Start: uint32(start),
		Size:  size,
	}
	mem.Init()

	start, buf := mem.Reserve(size, 0)
	defer mem.Release(start)

	if len(w) > 0 {
		copy(buf, w)
	} else {
		b = make([]byte, size)
		copy(b, buf)
	}

	return
}

func memReadCmd(_ *term.Terminal, arg []string) (res string, err error) {
	addr, err := strconv.ParseUint(arg[0], 16, 32)

	if err != nil {
		return "", fmt.Errorf("invalid address, %v", err)
	}

	size, err := strconv.ParseUint(arg[1], 10, 32)

	if err != nil {
		return "", fmt.Errorf("invalid size, %v", err)
	}

	if (addr%4) != 0 || (size%4) != 0 {
		return "", fmt.Errorf("only 32-bit aligned accesses are supported")
	}

	if size > maxBufferSize {
		return "", fmt.Errorf("size argument must be <= %d", maxBufferSize)
	}

	return hex.Dump(mem(uint32(addr), int(size), nil)), nil
}

func memWriteCmd(_ *term.Terminal, arg []string) (res string, err error) {
	addr, err := strconv.ParseUint(arg[0], 16, 32)

	if err != nil {
		return "", fmt.Errorf("invalid address, %v", err)
	}

	val, err := strconv.ParseUint(arg[1], 16, 32)

	if err != nil {
		return "", fmt.Errorf("invalid data, %v", err)
	}

	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(val))

	mem(uint32(addr), 4, buf)

	return
}

func memTest() {
	var memstats runtime.MemStats

	chunks := mathrand.Intn(chunksMax) + 1
	chunkSize := fillSize / chunks

	// Instead of forcing runtime.GC() as shown in the loop, gcpercent can
	// be tuned to a value sufficiently low to prevent the next GC target
	// being set beyond the end of available RAM. A lower than default
	// (100) value (such as 80 for this example) triggers GC more
	// frequently and avoids forced GC runs.
	//
	// This is not something unique to `GOOS=tamago` but more evident as,
	// when running on bare metal, there is no swap or OS virtual memory.
	//
	// gcpercent := 80
	// debug.SetGCPercent(gcpercent)

	msg("memory allocation (%d runs)", runs)

	for run := 1; run <= runs; run++ {
		log.Printf("allocating %d * %d MiB chunks (%d/%d)", chunks, chunkSize/(1024*1024), run, runs)

		buf := make([][]byte, chunks)

		for i := 0; i <= chunks-1; i++ {
			buf[i] = make([]byte, chunkSize)
		}

		// When getting close to the end of available RAM, the next GC
		// target might be set beyond it. Therfore in this specific
		// test it is best to force a GC run.
		//
		// This is not something unique to `GOOS=tamago` but more
		// evident as when running bare metal we have no swap or OS
		// virtual memory.
		runtime.GC()
	}

	runtime.ReadMemStats(&memstats)
	totalAllocated := uint64(runs) * uint64(chunks) * uint64(chunkSize)

	log.Printf("%d MiB allocated (Mallocs: %d Frees: %d HeapSys: %d NumGC:%d)",
		totalAllocated/(1024*1024), memstats.Mallocs, memstats.Frees, memstats.HeapSys, memstats.NumGC)
}
