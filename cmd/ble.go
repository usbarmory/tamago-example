// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build usbarmory

package cmd

import (
	"errors"
	"fmt"
	"io"
	"log"
	"runtime"

	"github.com/usbarmory/tamago-example/shell"
	usbarmory "github.com/usbarmory/tamago/board/usbarmory/mk2"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

const CR = 0x0d

func init() {
	shell.Add(shell.Cmd{
		Name: "ble",
		Help: "BLE serial console",
		Fn:   bleCmd,
	})
}

func bleCmd(console *shell.Interface, _ []string) (_ string, err error) {
	if !imx6ul.Native {
		return "", errors.New("unavailable under emulation")
	}

	if usbarmory.BLE.UART == nil {
		return "", errors.New("BLE module is not initialized")
	}

	log.Printf("switching to BLE console, type `quit` to exit")

	defer func() {
		log.Printf("resetting BLE module")
		usbarmory.BLE.Reset()
	}()

	t := console.Terminal

	t.SetPrompt(string(t.Escape.Blue) + "BLE> " + string(t.Escape.Reset))
	defer t.SetPrompt(string(t.Escape.Red) + "> " + string(t.Escape.Reset))

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

			fmt.Fprintf(t, "%s", string(c))
		}
	}()

	var tx string

	for {
		tx, err = t.ReadLine()

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
