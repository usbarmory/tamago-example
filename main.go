// https://github.com/usbarmory/tamago-example
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"

	"github.com/usbarmory/tamago-example/cmd"
	"github.com/usbarmory/tamago-example/network"
)

var Build string
var Revision string
var banner string

var exit chan bool

func init() {
	log.SetFlags(0)

	cmd.Banner = fmt.Sprintf("%s/%s (%s) • %s %s",
		runtime.GOOS, runtime.GOARCH, runtime.Version(),
		Revision, Build)

	cmd.Banner += fmt.Sprintf(" • %s", cmd.Target())
}

func main() {
	logFile, _ := os.OpenFile("/tamago-example.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	log.Println(banner)

	if cmd.Remote() {
		network.Start(logFile)
	} else {
		cmd.SerialConsole()
	}

	log.Println("Goodbye from tamago/arm")
}
