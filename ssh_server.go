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
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"runtime/debug"
	"runtime/pprof"
	"strconv"
	"unsafe"

	"github.com/f-secure-foundry/tamago/soc/imx6"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

const help = `
  help                                   # this help
  exit, quit                             # close session
  example                                # launch example test code
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

const MD_LIMIT = 102400

var LED func(string, bool) error
var i2c []*imx6.I2C

var dcpCommandPattern = regexp.MustCompile(`dcp (\d+) (\d+)`)
var otpCommandPattern = regexp.MustCompile(`otp (\d+) (\d+)`)
var ledCommandPattern = regexp.MustCompile(`led (white|blue) (on|off)`)
var mmcCommandPattern = regexp.MustCompile(`mmc (\d) ([[:xdigit:]]+) (\d+)`)
var i2cCommandPattern = regexp.MustCompile(`i2c (\d) ([[:xdigit:]]+) ([[:xdigit:]]+) (\d+)`)
var memoryCommandPattern = regexp.MustCompile(`(md|mw) ([[:xdigit:]]+) (\d+|[[:xdigit:]]+)`)

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

		buf := make([]byte, size)

		for i := 0; i < int(size); i += 4 {
			reg := (*uint32)(unsafe.Pointer(uintptr(addr + uint64(i))))
			val := *reg

			buf[i] = byte((val >> 24) & 0xff)
			buf[i+1] = byte((val >> 16) & 0xff)
			buf[i+2] = byte((val >> 8) & 0xff)
			buf[i+3] = byte(val & 0xff)
		}

		res = hex.Dump(buf)
	case "mw":
		val, err := strconv.ParseUint(arg[2], 16, 32)

		if err != nil {
			return fmt.Sprintf("invalid data: %v", err)
		}

		reg := (*uint32)(unsafe.Pointer(uintptr(addr)))
		*reg = uint32(val)
	}

	return
}

func handleCommand(term *term.Terminal, cmd string) (err error) {
	var res string

	switch cmd {
	case "exit", "quit":
		res = "logout"
		err = io.EOF
	case "example":
		example(false)
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

func handleChannel(newChannel ssh.NewChannel) {
	if t := newChannel.ChannelType(); t != "session" {
		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}

	conn, requests, err := newChannel.Accept()

	if err != nil {
		log.Printf("error accepting channel, %v", err)
		return
	}

	term := term.NewTerminal(conn, "")
	term.SetPrompt(string(term.Escape.Red) + "> " + string(term.Escape.Reset))

	go func() {
		defer conn.Close()

		log.SetOutput(io.MultiWriter(os.Stdout, term))
		defer log.SetOutput(os.Stdout)

		fmt.Fprintf(term, "%s\n", banner)
		fmt.Fprintf(term, "%s\n", string(term.Escape.Cyan)+help+string(term.Escape.Reset))

		for {
			cmd, err := term.ReadLine()

			if err == io.EOF {
				break
			}

			if err != nil {
				log.Printf("readline error: %v", err)
				continue
			}

			if cmd == "ble" {
				err = bleConsole(term)
			} else {
				err = handleCommand(term, cmd)
			}

			if err == io.EOF {
				break
			}
		}

		log.Printf("closing ssh connection")
	}()

	go func() {
		for req := range requests {
			reqSize := len(req.Payload)

			switch req.Type {
			case "shell":
				// do not accept payload commands
				if len(req.Payload) == 0 {
					req.Reply(true, nil)
				}
			case "pty-req":
				// p10, 6.2.  Requesting a Pseudo-Terminal, RFC4254
				if reqSize < 4 {
					log.Printf("malformed pty-req request")
					continue
				}

				termVariableSize := int(req.Payload[3])

				if reqSize < 4+termVariableSize+8 {
					log.Printf("malformed pty-req request")
					continue
				}

				w := binary.BigEndian.Uint32(req.Payload[4+termVariableSize:])
				h := binary.BigEndian.Uint32(req.Payload[4+termVariableSize+4:])

				log.Printf("resizing terminal (%s:%dx%d)", req.Type, w, h)
				term.SetSize(int(w), int(h))

				req.Reply(true, nil)
			case "window-change":
				// p10, 6.7.  Window Dimension Change Message, RFC4254
				if reqSize < 8 {
					log.Printf("malformed window-change request")
					continue
				}

				w := binary.BigEndian.Uint32(req.Payload)
				h := binary.BigEndian.Uint32(req.Payload[4:])

				log.Printf("resizing terminal (%s:%dx%d)", req.Type, w, h)
				term.SetSize(int(w), int(h))
			}
		}
	}()
}

func handleChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		go handleChannel(newChannel)
	}
}

func startSSHServer(s *stack.Stack, addr tcpip.Address, port uint16, nic tcpip.NICID) {
	var err error

	fullAddr := tcpip.FullAddress{Addr: addr, Port: port, NIC: nic}
	listener, err := gonet.ListenTCP(s, fullAddr, ipv4.ProtocolNumber)

	if err != nil {
		log.Fatal("listener error: ", err)
	}

	srv := &ssh.ServerConfig{
		NoClientAuth: true,
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	if err != nil {
		log.Fatal("private key generation error: ", err)
	}

	signer, err := ssh.NewSignerFromKey(key)

	if err != nil {
		log.Fatal("key conversion error: ", err)
	}

	log.Printf("starting ssh server (%s) at %s:%d", ssh.FingerprintSHA256(signer.PublicKey()), addr.String(), port)

	srv.AddHostKey(signer)

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Printf("error accepting connection, %v", err)
			continue
		}

		sshConn, chans, reqs, err := ssh.NewServerConn(conn, srv)

		if err != nil {
			log.Printf("error accepting handshake, %v", err)
			continue
		}

		log.Printf("new ssh connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())

		go ssh.DiscardRequests(reqs)
		go handleChannels(chans)
	}
}
