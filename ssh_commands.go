// https://github.com/f-secure-foundry/tamago-example
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"regexp"
	"runtime/debug"
	"runtime/pprof"
	"strconv"

	"golang.org/x/term"
)

const MD_LIMIT = 102400

const help = `
  help                                   # this help
  exit, quit                             # close session
  info                                   # SoC/board information
  test                                   # launch tests
  rand                                   # gather 32 bytes from TRNG
  reboot                                 # reset the SoC/board
  stack                                  # stack trace of current goroutine
  stackall                               # stack trace of all goroutines
  ble                                    # enter BLE serial console
  i2c <n> <hex slave> <hex addr> <size>  # I²C bus read
  mmc <n> <hex offset> <size>            # internal MMC/SD card read
  md  <hex offset> <size>                # memory display (use with caution)
  mw  <hex offset> <hex value>           # memory write   (use with caution)
  led (white|blue) (on|off)              # LED control
  dcp <size> <sec>                       # benchmark hardware encryption
  otp <bank> <word>                      # OTP fuse display
`

var dcpCommandPattern = regexp.MustCompile(`dcp (\d+) (\d+)`)
var otpCommandPattern = regexp.MustCompile(`otp (\d+) (\d+)`)
var ledCommandPattern = regexp.MustCompile(`led (white|blue) (on|off)`)
var mmcCommandPattern = regexp.MustCompile(`mmc (\d) ([[:xdigit:]]+) (\d+)`)
var i2cCommandPattern = regexp.MustCompile(`i2c (\d) ([[:xdigit:]]+) ([[:xdigit:]]+) (\d+)`)
var memoryCommandPattern = regexp.MustCompile(`(md|mw) ([[:xdigit:]]+) (\d+|[[:xdigit:]]+)`)

var LED func(string, bool) error

func dcpCommand(arg []string) (res string) {
	size, err := strconv.Atoi(arg[0])

	if err != nil {
		return fmt.Sprintf("invalid size: %v", err)
	}

	sec, err := strconv.Atoi(arg[1])

	if err != nil {
		return fmt.Sprintf("invalid duration: %v", err)
	}

	log.Printf("Doing aes-128 cbc for %ds on %d blocks", sec, size)

	n, d, err := testDecryption(size, sec)

	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf("%d aes-128 cbc's in %s", n, d)
}

func otpCommand(arg []string) (res string) {
	bank, err := strconv.Atoi(arg[0])

	if err != nil {
		return fmt.Sprintf("invalid bank: %v", err)
	}

	word, err := strconv.Atoi(arg[1])

	if err != nil {
		return fmt.Sprintf("invalid word: %v", err)
	}

	res, err = readOTP(bank, word)

	if err != nil {
		return err.Error()
	}

	return
}

func ledCommand(arg []string) (res string) {
	if LED == nil {
		return
	}

	name := arg[0]
	state := arg[1]

	if state == "on" {
		LED(name, true)
	} else {
		LED(name, false)
	}

	return
}

func mmcCommand(arg []string) (res string) {
	n, err := strconv.ParseUint(arg[0], 10, 8)

	if err != nil {
		return fmt.Sprintf("invalid card index: %v", err)
	}

	addr, err := strconv.ParseUint(arg[1], 16, 32)

	if err != nil {
		return fmt.Sprintf("invalid address: %v", err)
	}

	size, err := strconv.ParseUint(arg[2], 10, 32)

	if err != nil {
		return fmt.Sprintf("invalid size: %v", err)
	}

	if size > MD_LIMIT {
		return fmt.Sprintf("please only use a size argument <= %d", MD_LIMIT)
	}

	if len(cards) < int(n+1) {
		return "invalid card index"
	}

	buf, err := cards[n].Read(int64(addr), int64(size))

	if err != nil {
		return err.Error()
	}

	return hex.Dump(buf)
}

func i2cCommand(arg []string) (res string) {
	n, err := strconv.ParseUint(arg[0], 10, 8)

	if err != nil {
		return fmt.Sprintf("invalid bus index: %v", err)
	}

	slave, err := strconv.ParseUint(arg[1], 16, 7)

	if err != nil {
		return fmt.Sprintf("invalid slave: %v", err)
	}

	addr, err := strconv.ParseUint(arg[2], 16, 32)

	if err != nil {
		return fmt.Sprintf("invalid address: %v", err)
	}

	size, err := strconv.ParseUint(arg[3], 10, 32)

	if err != nil {
		return fmt.Sprintf("invalid size: %v", err)
	}

	if size > MD_LIMIT {
		return fmt.Sprintf("please only use a size argument <= %d", MD_LIMIT)
	}

	if n <= 0 || len(i2c) < int(n) {
		return "invalid bus index"
	}

	buf, err := i2c[n-1].Read(uint8(slave), uint32(addr), 1, int(size))

	if err != nil {
		return err.Error()
	}

	return hex.Dump(buf)
}

func memoryCommand(arg []string) (res string) {
	addr, err := strconv.ParseUint(arg[1], 16, 32)

	if err != nil {
		return fmt.Sprintf("invalid address: %v", err)
	}

	switch arg[0] {
	case "md":
		size, err := strconv.ParseUint(arg[2], 10, 32)

		if err != nil {
			return fmt.Sprintf("invalid size: %v", err)
		}

		if (addr%4) != 0 || (size%4) != 0 {
			return "please only perform 32-bit aligned accesses"
		}

		if size > MD_LIMIT {
			return fmt.Sprintf("please only use a size argument <= %d", MD_LIMIT)
		}

		return hex.Dump(mem(uint32(addr), int(size), nil))
	case "mw":
		val, err := strconv.ParseUint(arg[2], 16, 32)

		if err != nil {
			return fmt.Sprintf("invalid data: %v", err)
		}

		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, uint32(val))

		mem(uint32(addr), 4, buf)
	}

	return
}

func handleCommand(term *term.Terminal, cmd string) (err error) {
	var res string

	switch cmd {
	case "exit", "quit":
		res = "logout"
		err = io.EOF
	case "info":
		res = info()
	case "test":
		test(false)
	case "help":
		res = string(term.Escape.Cyan) + help + string(term.Escape.Reset)
	case "rand":
		buf := make([]byte, 32)
		rand.Read(buf)
		res = string(term.Escape.Cyan) + fmt.Sprintf("%x", buf) + string(term.Escape.Reset)
	case "reboot":
		reset()
	case "stack":
		res = string(debug.Stack())
	case "stackall":
		buf := new(bytes.Buffer)
		pprof.Lookup("goroutine").WriteTo(buf, 1)
		res = buf.String()
	default:
		if m := dcpCommandPattern.FindStringSubmatch(cmd); len(m) == 3 {
			res = dcpCommand(m[1:])
		} else if m := otpCommandPattern.FindStringSubmatch(cmd); len(m) == 3 {
			res = otpCommand(m[1:])
		} else if m := ledCommandPattern.FindStringSubmatch(cmd); len(m) == 3 {
			res = ledCommand(m[1:])
		} else if m := mmcCommandPattern.FindStringSubmatch(cmd); len(m) == 4 {
			res = mmcCommand(m[1:])
		} else if m := i2cCommandPattern.FindStringSubmatch(cmd); len(m) == 5 {
			res = i2cCommand(m[1:])
		} else if m := memoryCommandPattern.FindStringSubmatch(cmd); len(m) == 4 {
			res = memoryCommand(m[1:])
		} else {
			res = "unknown command, type `help`"
		}
	}

	fmt.Fprintln(term, res)

	return
}
