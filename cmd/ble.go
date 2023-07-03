// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build usbarmory
// +build usbarmory

package cmd

import (
	"errors"
	"fmt"
	"io"
	"log"
	"runtime"

	"golang.org/x/term"

	usbarmory "github.com/usbarmory/tamago/board/usbarmory/mk2"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

const CR = 0x0d

func init() {
	Add(Cmd{
		Name: "ble",
		Help: "BLE serial console",
		Fn:   bleCmd,
	})
}

func bleCmd(_ *Interface, term *term.Terminal, _ []string) (_ string, err error) {
	if !imx6ul.Native {
		return "", errors.New("unsupported under emulation")
	}

	if usbarmory.BLE.UART == nil {
		return "", errors.New("BLE module is not initialized")
	}

	log.Printf("switching to BLE console, type `quit` to exit")

	defer func() {
		log.Printf("resetting BLE module")
		usbarmory.BLE.Reset()
	}()

	term.SetPrompt(string(term.Escape.Blue) + "BLE> " + string(term.Escape.Reset))
	defer term.SetPrompt(string(term.Escape.Red) + "> " + string(term.Escape.Reset))

	exit := make(chan bool)

	go func() {
		for {
			select {
			case <-exit:
				return
			default:
			}

			c, valid := usbarmory.UART1.Rx()

			if !valid || c == CR {
				runtime.Gosched()
				continue
			}

			fmt.Fprintf(term, "%s", string(c))
		}
	}()

	var tx string

	for {
		tx, err = term.ReadLine()

		if err == io.EOF {
			continue
		} else if err != nil {
			break
		}

		if tx == "quit" {
			break
		}

		usbarmory.BLE.UART.Write([]byte(tx + "\r"))
	}

	exit <- true

	return
}
