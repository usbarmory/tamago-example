// https://github.com/f-secure-foundry/tamago-example
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"log"
	"runtime"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/f-secure-foundry/tamago/imx6"
	"github.com/f-secure-foundry/tamago/usbarmory/mark-two"
)

const CR = 0x0d

func bleConsole(term *terminal.Terminal) (err error) {
	log.Printf("switching to BLE console, type `quit` to exit")

	if usbarmory.BLE.UART == nil {
		log.Printf("BLE module is not initialized")
		return io.EOF
	}

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

			c, valid := imx6.UART1.Rx()

			if !valid || c == CR {
				runtime.Gosched()
				continue
			}

			fmt.Fprintf(term, "%s", string(c))
		}
	}()

	for {
		var tx string

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
