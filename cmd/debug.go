// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build debug

package cmd

import (
	"fmt"
	"regexp"

	"github.com/usbarmory/tamago-example/shell"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

func init() {
	shell.Add(shell.Cmd{
		Name:    "go",
		Args:    1,
		Pattern: regexp.MustCompile(`^go (.*)`),
		Syntax:  "<expr>",
		Help:    "Go stdlib eval (go:build debug)",
		Fn:      replCmd,
	})
}

func replCmd(console *shell.Interface, arg []string) (_ string, err error) {
	i := interp.New(interp.Options{
		Unrestricted: false,
	})

	if err = i.Use(stdlib.Symbols); err != nil {
		return
	}

	if err = i.Use(interp.Symbols); err != nil {
		return
	}

	i.ImportUsed()

	v, err := i.Eval(arg[0])

	return fmt.Sprintf("%+v", v), err
}
