// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package shell implements a terminal console handler for user defined
// commands.
package shell

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
	"text/tabwriter"

	"golang.org/x/term"
)

// CmdFn represents a command handler.
type CmdFn func(iface *Interface, term *term.Terminal, arg []string) (res string, err error)

// Cmd represents a shell command.
type Cmd struct {
	// Name is the command name.
	Name string
	// Args defines the number of command arguments, meant to be in the
	// Pattern capturing brackets.
	Args int
	// Pattern defines the command syntax and arguments.
	Pattern *regexp.Regexp
	// Syntax defines the Help() command syntax field.
	Syntax string
	// Help defines the Help() command description field.
	Help string
	// Fn defines the command handler.
	Fn CmdFn
}

var cmds = make(map[string]*Cmd)

// Interface represents a terminal interface.
type Interface struct {
	// Banner represents the welcome message
	Banner string
	// Log represents the interface log file
	Log *os.File
}

// Add registers a terminal interface command.
func Add(cmd Cmd) {
	cmds[cmd.Name] = &cmd
}

// Help returns a formatted string with instructions for all registered
// commands.
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

// Exec executes a terminal command.
func (iface *Interface) Exec(term *term.Terminal, cmd []byte) {
	if err := iface.handle(term, string(cmd)); err != nil {
		fmt.Fprintf(term, "command error (%s), %v\n", cmd, err)
	}
}

// Terminal handles terminal input.
func (iface *Interface) Terminal(term *term.Terminal) {
	term.SetPrompt(string(term.Escape.Red) + "> " + string(term.Escape.Reset))

	fmt.Fprintf(term, "\n%s\n\n", iface.Banner)
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

// StartTerminal starts a terminal on a shell interface handler.
func StartTerminal(iface *Interface, terminal io.ReadWriter) {
	term := term.NewTerminal(terminal, "")
	term.SetPrompt(string(term.Escape.Red) + "> " + string(term.Escape.Reset))

	iface.Terminal(term)
}
