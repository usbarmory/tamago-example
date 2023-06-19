// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package reg provides primitives for retrieving and modifying hardware
// registers.
//
// This package is only meant to be used with `GOOS=tamago` as supported by the
// TamaGo framework for bare metal Go on ARM/RISC-V SoCs, see
// https://github.com/usbarmory/tamago.
package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"github.com/usbarmory/tamago/board/usbarmory/mk2"
)

func Get(addr uint32, pos int, mask int) uint32 {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))
	r := atomic.LoadUint32(reg)

	return uint32((int(r) >> pos) & mask)
}

func Set(addr uint32, pos int) {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))

	r := atomic.LoadUint32(reg)
	r |= (1 << pos)

	atomic.StoreUint32(reg, r)
}

func Clear(addr uint32, pos int) {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))

	r := atomic.LoadUint32(reg)
	r &= ^(1 << pos)

	atomic.StoreUint32(reg, r)
}

func SetTo(addr uint32, pos int, val bool) {
	if val {
		Set(addr, pos)
	} else {
		Clear(addr, pos)
	}
}

func SetN(addr uint32, pos int, mask int, val uint32) {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))

	r := atomic.LoadUint32(reg)
	r = (r & (^(uint32(mask) << pos))) | (val << pos)

	atomic.StoreUint32(reg, r)
}

func ClearN(addr uint32, pos int, mask int) {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))

	r := atomic.LoadUint32(reg)
	r &= ^(uint32(mask) << pos)

	atomic.StoreUint32(reg, r)
}

// defined in reg32_*.s
// func Move(dst uint32, src uint32)

func Read(addr uint32) uint32 {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))
	return atomic.LoadUint32(reg)
}

func Write(addr uint32, val uint32) {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))
	atomic.StoreUint32(reg, val)
}

func WriteBack(addr uint32) {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))

	r := atomic.LoadUint32(reg)
	r |= r

	atomic.StoreUint32(reg, r)
}

func Or(addr uint32, val uint32) {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))

	r := atomic.LoadUint32(reg)
	r |= val

	atomic.StoreUint32(reg, r)
}

// Wait waits for a specific register bit to match a value. This function
// cannot be used before runtime initialization with `GOOS=tamago`.
func Wait(addr uint32, pos int, mask int, val uint32) {
	for Get(addr, pos, mask) != val {
		// tamago is single-threaded, give other goroutines a chance
		runtime.Gosched()
	}
}

// WaitFor waits, until a timeout expires, for a specific register bit to match
// a value. The return boolean indicates whether the wait condition was checked
// (true) or if it timed out (false). This function cannot be used before
// runtime initialization.
func WaitFor(timeout time.Duration, addr uint32, pos int, mask int, val uint32) bool {
	start := time.Now()

	for Get(addr, pos, mask) != val {
		// tamago is single-threaded, give other goroutines a chance
		runtime.Gosched()

		if time.Since(start) >= timeout {
			return false
		}
	}

	return true
}

// WaitSignal waits, until a channel is closed, for a specific register bit to
// match a value. The return boolean indicates whether the wait condition was
// checked (true) or cancelled (false). This function cannot be used before
// runtime initialization.
func WaitSignal(done chan bool, addr uint32, pos int, mask int, val uint32) bool {
	for Get(addr, pos, mask) != val {
		// tamago is single-threaded, give other goroutines a chance
		runtime.Gosched()

		select {
		case <-done:
			return false
		default:
		}
	}

	return true
}

// reg32File represents a single, 32-bit-aligned, 32-bit-sized, register.
type reg32File struct {
	offset uint32
}

func init() {
	log.Printf("=======> t9 device reg32")
	err := syscall.MkDev("/dev/reg32", 0666, openReg32)
	log.Printf("err %v", err)
}

func openReg32() (syscall.DevFile, error) {
	return reg32File{}, nil
}

func (f reg32File) close() error {
	return nil
}

var ErrNotAligned = errors.New("Not aligned")

func check(size int, offset int64) error {
	// The offset must be aligned x4
	if (offset & 3) != 0 {
		return fmt.Errorf("offset %#x: %w", ErrNotAligned)
	}
	// The size must be aligned x4
	if (size & 3) != 0 {
		return fmt.Errorf("size %#x: %w", ErrNotAligned)
	}
	return nil
}

func (f reg32File) Pread(b []byte, offset int64) (int, error) {
	if err := check(len(b), offset); err != nil {
		return -1, err
	}
	var longs = make([]uint32, len(b)/4)
	// read them from the registers ...
	binary.Write(bytes.NewBuffer(b), binary.LittleEndian, longs)

	return len(b), nil
}

func (f reg32File) Pwrite(b []byte, offset int64) (int, error) {
	if err := check(len(b), offset); err != nil {
		return -1, err
	}
	var longs = make([]uint32, len(b)/4)
	binary.Read(bytes.NewBuffer(b), binary.LittleEndian, longs)
	return len(b), nil
}

type ledFile struct {
	f         reg32File
	color     string
	onoff     string
	lasterror error
}

var leds = []*ledFile{
	&ledFile{color: "white", onoff: "on", f: reg32File{offset: 0}},
	&ledFile{color: "blue", onoff: "on", f: reg32File{offset: 0}},
}

func init() {
	syscall.MkDev("/dev/white", 0666, openwhite)
	syscall.MkDev("/dev/blue", 0666, openblue)
}

func openwhite() (syscall.DevFile, error) {
	return leds[0], nil
}

func openblue() (syscall.DevFile, error) {
	return leds[1], nil
}

func (f ledFile) Close() error {
	return nil
}

func (f ledFile) Pread(b []byte, offset int64) (int, error) {
	n, err := bytes.NewReader([]byte(fmt.Sprintf(`{"color": %q,"state": %q, "lasterror": %v}`, f.color, f.onoff, f.lasterror))).ReadAt(b, offset)
	if err == io.EOF && n > 0 {
		err = nil
	}
	return n, err
}

func (f ledFile) Pwrite(b []byte, offset int64) (int, error) {
	f.lasterror = fmt.Errorf("%v: %q", f, string(b))
	cmd := strings.Fields(string(b))
	if len(cmd) != 1 {
		f.lasterror = fmt.Errorf("%q:%q usage: on|off", f.color, string(b))
		return -1, fmt.Errorf("usage: blue|white on|off")
	}
	var onoff bool
	switch cmd[0] {
	case "on":
		onoff = true
	case "off":
	default:
		f.lasterror = fmt.Errorf("%q:%q usage: on|off", f.color, string(b))
		log.Printf("%q:%q usage: on|off", f.color, string(b))
		return -1, fmt.Errorf("%q:%q usage: on|off", f.color, string(b))
	}

	err := mk2.LED(f.color, onoff)

	if err != nil {
		log.Printf("%q:%q mk2.LED: %v", f.color, string(b), err)
		f.lasterror = fmt.Errorf("%q:%q mk2.LED: %v", f.color, string(b), err)
	}
	return len(b), err
}
