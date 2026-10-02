// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
package cpu

import (
	"fmt"
	"io"
	"runtime"
	"time"
)

type CPU interface {
	IdleTime() int64
	GetTime() int64
}

type Sample struct {
	Time int64
	Idle int64
}

func Read(cpu CPU) Sample {
	return Sample{
		Idle: cpu.IdleTime(),
		Time: cpu.GetTime(),
	}
}

func Load(a, b Sample) float64 {
	elapsed := b.Time - a.Time

	if elapsed <= 0 {
		return 0
	}

	idle := max(b.Idle-a.Idle, 0)

	if idle > elapsed {
		idle = elapsed
	}

	return 100 * (1 - float64(idle)/float64(elapsed))
}

func Measure(cpu CPU, d time.Duration) float64 {
	prev := Read(cpu)
	time.Sleep(d)
	return Load(prev, Read(cpu))
}

func Top(cpu CPU, interval time.Duration, n int, out io.Writer) {
	var m runtime.MemStats

	prev := Read(cpu)

	for i := 0; n <= 0 || i < n; i++ {
		time.Sleep(interval)

		cur := Read(cpu)
		busy := Load(prev, cur)
		prev = cur

		runtime.ReadMemStats(&m)

		fmt.Fprintf(out, "%5.1f%% busy, %5.1f%% idle   Goroutines: %-4d  Heap: %d KiB\n",
			busy, 100-busy, runtime.NumGoroutine(), m.HeapAlloc/1024)
	}
}
