// https://github.com/usbarmory/tamago-example
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"

	"golang.org/x/term"
)

const maxBufferSize = 102400

const (
	separator = "-"
	separatorSize = 80
)

type CmdFn func(term *term.Terminal, arg []string) (res string, err error)

type Cmd struct {
	Name    string
	Args    int
	Pattern *regexp.Regexp
	Syntax  string
	Help    string
	Fn      CmdFn
}

var Banner string
var cmds = make(map[string]*Cmd)
var console io.ReadWriter

func Add(cmd Cmd) {
	cmds[cmd.Name] = &cmd
}

func msg(format string, args ...interface{}) {
	s := strings.Repeat(separator, 2) + " "
	s += fmt.Sprintf(format, args...)
	s += strings.Repeat(separator, separatorSize - len(s))

	log.Println(s)
}

func Help(term *term.Terminal) string {
	var help bytes.Buffer
	var names []string

	t := tabwriter.NewWriter(&help, 16, 8, 0, '\t', tabwriter.TabIndent)

	for name, _ := range cmds {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		_, _ = fmt.Fprintf(t, "%s\t%s\t # %s\n", cmds[name].Name, cmds[name].Syntax, cmds[name].Help)
	}

	_ = t.Flush()

	return string(term.Escape.Cyan) + help.String() + string(term.Escape.Reset)
}

func Handle(term *term.Terminal, line string) (err error) {
	var match *Cmd
	var arg []string
	var res string

	for _, cmd := range cmds {
		if cmd.Pattern == nil {
			if cmd.Name == line {
				match = cmd
				break
			}
		} else if m := cmd.Pattern.FindStringSubmatch(line); len(m) > 0  && (len(m) - 1 == cmd.Args) {
			match = cmd
			arg = m[1:]
			break
		}
	}

	if match == nil {
		return errors.New("unknown command, type `help`")
	}

	if res, err = match.Fn(term, arg); err != nil {
		return
	}

	fmt.Fprintln(term, res)

	return
}

func Console(term *term.Terminal) {
	fmt.Fprintf(term, "%s\n\n", Banner)
	fmt.Fprintf(term, "%s\n", Help(term))

	for {
		s, err := term.ReadLine()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Printf("readline error: %v", err)
			continue
		}

		if err = Handle(term, s); err != nil {
			if err == io.EOF {
				break
			}

			log.Printf("command error: %v", err)
		}
	}
}

func SerialConsole() {
	term := term.NewTerminal(console, "")
	term.SetPrompt(string(term.Escape.Red) + "> " + string(term.Escape.Reset))

	Console(term)
}
