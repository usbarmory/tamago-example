// https://github.com/usbarmory/tamago-example
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build usbarmory
// +build usbarmory

package main

import (
	"fmt"
	"io"
	"log"
	"runtime"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	usbarmory "github.com/usbarmory/tamago/board/usbarmory/mk2"
	"github.com/usbarmory/tamago/soc/imx6"
)

const CR = 0x0d

var boardName = "USB armory Mk II"

func init() {
	i2c = append(i2c, imx6.I2C1)

	cards = append(cards, usbarmory.SD)
	cards = append(cards, usbarmory.MMC)

	LED = usbarmory.LED

	if imx6.Native && (imx6.Family == imx6.IMX6UL || imx6.Family == imx6.IMX6ULL) {
		boardName = usbarmory.Model()

		// On the USB armory Mk II the standard serial console (UART2) is
		// exposed through the debug accessory, which needs to be enabled.
		debugConsole, _ := usbarmory.DetectDebugAccessory(250 * time.Millisecond)
		<-debugConsole

		log.Println("-- i.mx6 ble ---------------------------------------------------------")
		usbarmory.BLE.Init()
		log.Println("ANNA-B112 BLE module initialized")
	}
}

func reset() {
	usbarmory.Reset()
}

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
