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
	"os"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"

	"golang.org/x/term"
)

const maxBufferSize = 102400

const (
	separator     = "-"
	separatorSize = 80
)

type CmdFn func(iface *Interface, term *term.Terminal, arg []string) (res string, err error)

type Cmd struct {
	Name    string
	Args    int
	Pattern *regexp.Regexp
	Syntax  string
	Help    string
	Fn      CmdFn
}

type Interface struct {
	Log *os.File
}

var Banner string
var cmds = make(map[string]*Cmd)
var uart io.ReadWriter

func Add(cmd Cmd) {
	cmds[cmd.Name] = &cmd
}

func msg(format string, args ...interface{}) {
	s := strings.Repeat(separator, 2) + " "
	s += fmt.Sprintf(format, args...)
	s += " " + strings.Repeat(separator, separatorSize-len(s))

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

func (iface *Interface) handle(term *term.Terminal, line string) (err error) {
	var match *Cmd
	var arg []string
	var res string

	for _, cmd := range cmds {
		if cmd.Pattern == nil {
			if cmd.Name == line {
				match = cmd
				break
			}
		} else if m := cmd.Pattern.FindStringSubmatch(line); len(m) > 0 && (len(m)-1 == cmd.Args) {
			match = cmd
			arg = m[1:]
			break
		}
	}

	if match == nil {
		return errors.New("unknown command, type `help`")
	}

	if res, err = match.Fn(iface, term, arg); err != nil {
		return
	}

	fmt.Fprintln(term, res)

	return
}

func (iface *Interface) LogFile() *os.File {
	return iface.Log
}

func (iface *Interface) Exec(term *term.Terminal, cmd []byte) {
	if err := iface.handle(term, string(cmd)); err != nil {
		fmt.Fprintf(term, "command error (%s), %v\n", cmd, err)
	}
}

func (iface *Interface) Terminal(term *term.Terminal) {
	term.SetPrompt(string(term.Escape.Red) + "> " + string(term.Escape.Reset))

	fmt.Fprintf(term, "\n%s\n\n", Banner)
	fmt.Fprintf(term, "%s\n", Help(term))

	for {
		s, err := term.ReadLine()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Printf("readline error, %v", err)
			continue
		}

		if err = iface.handle(term, s); err != nil {
			if err == io.EOF {
				break
			}

			fmt.Fprintf(term, "command error, %v\n", err)
		}
	}
}

func SerialConsole(iface *Interface) {
	term := term.NewTerminal(uart, "")
	term.SetPrompt(string(term.Escape.Red) + "> " + string(term.Escape.Reset))

	iface.Terminal(term)
}
