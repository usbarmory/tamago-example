// Copyright (c) WithSecure Corporation
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
	"github.com/usbarmory/tamago-example/internal/semihosting"
	"github.com/usbarmory/tamago-example/network"
	"github.com/usbarmory/tamago-example/shell"
)

func main() {
	log.SetFlags(0)

	logFile, _ := os.OpenFile("/tamago-example.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))

	name, _ := cmd.Target()

	banner := fmt.Sprintf("%s/%s (%s) • %s",
		runtime.GOOS, runtime.GOARCH, runtime.Version(), name)

	console := &shell.Interface{
		Banner: banner,
		Log:    logFile,
	}

	if hasUSB, hasEth := cmd.HasNetwork(); hasUSB || hasEth {
		network.SetupStaticWebAssets(banner)
		network.Init(console, hasUSB, hasEth, &cmd.NIC)
	}

	console.ReadWriter = cmd.Terminal
	console.Start(true)

	if runtime.GOARCH != "amd64" {
		semihosting.Exit()
	}

	runtime.Exit(0)
}
