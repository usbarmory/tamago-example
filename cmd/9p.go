// Copyright (c) 2023 The Ninep Authors. All rights reserved.

// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:

//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * The names of Ninep's contributors may not be used to endorse
// or promote products derived from this software without specific prior
// written permission.

// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//
// UFS is a userspace server which exports a filesystem over 9p2000.
//
// By default, it will export / over a TCP on port 5640 under the username
// of "harvey".

package cmd

import (
	"log"
	"net"

	"golang.org/x/term"

	ufs "github.com/Harvey-OS/ninep/filesystem"
	"github.com/Harvey-OS/ninep/protocol"

	"github.com/usbarmory/tamago-example/network"
	"github.com/usbarmory/tamago-example/shell"
)

func init() {
	shell.Add(shell.Cmd{
		Name: "9p",
		Help: "start 9p remote file server",
		Fn:   ninepCmd,
	})
}

func ninepCmd(iface *shell.Interface, _ *term.Terminal, _ []string) (_ string, err error) {
	log.Printf("starting 9p remote filesystem server")
	log.Printf("access with: `mount -t 9p -o trans=tcp,noextend %s <path>`", network.IP)

	listener9p, err := net.Listen("tcp", ":564")

	if err != nil {
		return
	}

	ufslistener, err := ufs.NewUFS(func(l *protocol.Listener) error {
		return nil
	})

	if err != nil {
		return
	}

	go func() {
		if err := ufslistener.Serve(listener9p); err != nil {
			log.Printf("9p server error: %v", err)
		}
	}()

	return
}
