// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package shell

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"text/tabwriter"
)

// CmdFn represents a command handler.
type CmdFn func(c *Interface, arg []string) (res string, err error)

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

// Add registers a terminal interface command.
func Add(cmd Cmd) {
	cmds[cmd.Name] = &cmd
}

// Confirm displays the argument prompt and waits for a "y" or "n" answer which
// is converted as return value.
func (c *Interface) Confirm(msg string) bool {
	if c.Terminal == nil {
		return false
	}

	c.Terminal.SetPrompt(msg)
	defer c.Terminal.SetPrompt(string(c.Terminal.Escape.Red) + c.Prompt + string(c.Terminal.Escape.Reset))

	input, err := c.Terminal.ReadLine()

	if err != nil {
		return false
	}

	return input == "y"
}

// Help returns a formatted string with instructions for all registered
// commands.
func Help(c *Interface, _ []string) (_ string, _ error) {
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
	res := help.String()

	if c.Terminal != nil {
		res = string(c.Terminal.Escape.Cyan) + res + string(c.Terminal.Escape.Reset)
	}

	fmt.Fprintln(c.Output, res)

	return
}
