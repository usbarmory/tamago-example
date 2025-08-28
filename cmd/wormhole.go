// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build mx6ullevk || usbarmory || cloud_hypervisor || microvm || firecracker

package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/psanford/wormhole-william/wormhole"

	"github.com/usbarmory/tamago-example/shell"
)

func init() {
	shell.Add(shell.Cmd{
		Name:    "wormhole",
		Args:    2,
		Pattern: regexp.MustCompile(`^wormhole (send|receive|recv) (.*)$`),
		Syntax:  "(send <path>|recv <code>)",
		Help:    "transfer file through magic wormhole",
		Fn:      wormholeCmd,
	})
}

func wormholeCmd(console *shell.Interface, arg []string) (res string, err error) {
	ctx := context.Background()
	client := &wormhole.Client{}

	switch arg[0] {
	case "send":
		f, err := os.Open(arg[1])

		if err != nil {
			return "", err
		}

		code, status, err := client.SendFile(ctx, arg[1], f)

		if err != nil {
			return "", err
		}

		fmt.Fprintf(console.Output, "on the other end of the wormhole please run recv with code %s\n", code)

		s := <-status

		if s.Error != nil {
			return "", s.Error
		} else if s.OK {
			fmt.Fprintln(console.Output, "file sent")
		} else {
			return "", errors.New("internal error")
		}
	case "recv", "receive":
		fileInfo, err := client.Receive(ctx, arg[1])

		if err != nil {
			return "", err
		}

		fmt.Fprintf(console.Output, "receiving %s (%d bytes)\n", fileInfo.Name, fileInfo.UncompressedBytes)

		file, err := os.OpenFile(fileInfo.Name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)

		if err != nil {
			return "", err
		}

		_, err = io.Copy(file, fileInfo)

		if err != nil {
			return "", err
		}
	}

	return
}
