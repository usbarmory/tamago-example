// https://github.com/f-secure-foundry/tamago-example
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"log"
	"runtime"

	"github.com/f-secure-foundry/tamago/arm"
	"github.com/f-secure-foundry/tamago/dma"
	"github.com/f-secure-foundry/tamago/soc/imx6"
)

func mem(start uint32, size int, w []byte) (b []byte) {
	// temporarily map page zero if required
	if z := uint32(1 << 20); start < z {
		imx6.ARM.ConfigureMMU(0, z, (arm.TTE_AP_001<<10)|arm.TTE_SECTION)
		defer imx6.ARM.ConfigureMMU(0, z, 0)
	}

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

func testAlloc(runs int, chunks int, chunkSize int) {
	var memstats runtime.MemStats

	// Instead of forcing runtime.GC() as shown in the loop, gcpercent can
	// be tuned to a value sufficiently low to prevent the next GC target
	// being set beyond the end of available RAM. A lower than default
	// (100) value (such as 80 for this example) triggers GC more
	// frequently and avoids forced GC runs.
	//
	// This is not something unique to `GOOS=tamago` but more evident as,
	// when running on bare metal, there is no swap or OS virtual memory.
	//
	//gcpercent := 80
	//debug.SetGCPercent(gcpercent)

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
