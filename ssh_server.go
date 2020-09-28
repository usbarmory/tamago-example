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
	"golang.org/x/crypto/ssh/terminal"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

const help = `
  help                              # this help
  exit, quit                        # close session
  example                           # launch example test code
  rand                              # gather 32 bytes from TRNG via crypto/rand
  reboot                            # reset watchdog timer
  stack                             # stack trace of current goroutine
  stackall                          # stack trace of all goroutines
  ble                               # enter BLE serial console
  mmc read <n> <hex offset> <size>  # internal MMC/SD card read
  md       <hex offset> <size>      # memory display (use with caution)
  mw       <hex offset> <hex value> # memory write   (use with caution)
  led      (white|blue) (on|off)    # LED control
  dcp      <size> <sec>             # benchmark hardware encryption
`

const MD_LIMIT = 102400

var LED func(string, bool) error

var dcpCommandPattern = regexp.MustCompile(`dcp (\d+) (\d+).*`)
var ledCommandPattern = regexp.MustCompile(`led (white|blue) (on|off).*`)
var cardCommandPattern = regexp.MustCompile(`mmc read (\d) ?([[:xdigit:]]+) (\d+|[[:xdigit:]]+).*`)
var memoryCommandPattern = regexp.MustCompile(`(md|mw) ?([[:xdigit:]]+) (\d+|[[:xdigit:]]+).*`)

func dcpCommand(arg1 string, arg2 string) (res string) {
	size, err := strconv.Atoi(arg1)

	if err != nil {
		return fmt.Sprintf("invalid size: %v", err)
	}

	sec, err := strconv.Atoi(arg2)

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

func ledCommand(name string, state string) (res string) {
	if LED == nil {
		return
	}

	if state == "on" {
		LED(name, true)
	} else {
		LED(name, false)
	}

	return
}

func cardCommand(arg1 string, arg2 string, arg3 string) (res string) {
	n, err := strconv.ParseUint(arg1, 10, 8)

	if err != nil {
		return fmt.Sprintf("invalid index: %v", err)
	}

	addr, err := strconv.ParseUint(arg2, 16, 32)

	if err != nil {
		return fmt.Sprintf("invalid address: %v", err)
	}

	val, err := strconv.ParseUint(arg3, 10, 32)

	if err != nil {
		return fmt.Sprintf("invalid size: %v", err)
	}

	if val > MD_LIMIT {
		return fmt.Sprintf("please only use a size argument <= %d", MD_LIMIT)
	}

	if len(cards) < int(n+1) {
		return "invalid index"
	}

	data, err := cards[n].Read(int64(addr), int64(val))

	if err != nil {
		return err.Error()
	}

	return hex.Dump(data)
}

func memoryCommand(op string, arg1 string, arg2 string) (res string) {
	var err error
	var val uint64
	var data []byte

	addr, err := strconv.ParseUint(arg1, 16, 32)

	if err != nil {
		return fmt.Sprintf("invalid address: %v", err)
	}

	switch op {
	case "md":
		val, err = strconv.ParseUint(arg2, 10, 32)

		if err != nil {
			return fmt.Sprintf("invalid size: %v", err)
		}

		if (addr%4) != 0 || (val%4) != 0 {
			return "please only perform 32-bit aligned accesses"
		}

		if val > MD_LIMIT {
			return fmt.Sprintf("please only use a size argument <= %d", MD_LIMIT)
		}

		data = make([]byte, val)

		for i := 0; i < int(val); i += 4 {
			reg := (*uint32)(unsafe.Pointer(uintptr(addr + uint64(i))))
			val := *reg

			data[i] = byte((val >> 24) & 0xff)
			data[i+1] = byte((val >> 16) & 0xff)
			data[i+2] = byte((val >> 8) & 0xff)
			data[i+3] = byte(val & 0xff)
		}

		res = hex.Dump(data)
	case "mw":
		val, err = strconv.ParseUint(arg2, 16, 32)

		if err != nil {
			return fmt.Sprintf("invalid data: %v", err)
		}

		reg := (*uint32)(unsafe.Pointer(uintptr(addr)))
		*reg = uint32(val)
	}

	return
}

func handleCommand(term *terminal.Terminal, cmd string) (err error) {
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
		imx6.Reboot()
	case "stack":
		res = string(debug.Stack())
	case "stackall":
		buf := new(bytes.Buffer)
		pprof.Lookup("goroutine").WriteTo(buf, 1)
		res = buf.String()
	default:
		if m := dcpCommandPattern.FindStringSubmatch(cmd); len(m) == 3 {
			res = dcpCommand(m[1], m[2])
		} else if m := ledCommandPattern.FindStringSubmatch(cmd); len(m) == 3 {
			res = ledCommand(m[1], m[2])
		} else if m := cardCommandPattern.FindStringSubmatch(cmd); len(m) == 4 {
			res = cardCommand(m[1], m[2], m[3])
		} else if m := memoryCommandPattern.FindStringSubmatch(cmd); len(m) == 4 {
			res = memoryCommand(m[1], m[2], m[3])
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

	term := terminal.NewTerminal(conn, "")
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
		log.Fatal("ECDSA key error: ", err)
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
