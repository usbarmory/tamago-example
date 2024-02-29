// https://github.com/usbarmory/tamago-example
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package network

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

type sshHandler interface {
	Exec(term *term.Terminal, cmd []byte)
	Terminal(term *term.Terminal)
	LogFile() *os.File
}

func handleTerminal(conn ssh.Channel, term *term.Terminal, handler sshHandler) {
	log.SetOutput(io.MultiWriter(os.Stdout, handler.LogFile(), term))
	defer log.SetOutput(io.MultiWriter(os.Stdout))

	handler.Terminal(term)

	log.Printf("closing ssh connection")
	conn.Close()
}

func handleChannel(newChannel ssh.NewChannel, handler sshHandler) {
	if t := newChannel.ChannelType(); t != "session" {
		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}

	conn, requests, err := newChannel.Accept()

	if err != nil {
		log.Printf("error accepting channel, %v", err)
		return
	}

	term := term.NewTerminal(conn, "")

	go func() {
		for req := range requests {
			reqSize := len(req.Payload)

			switch req.Type {
			case "exec":
				handler.Exec(term, req.Payload[4:])
				conn.Close()
				return
			case "shell":
				go handleTerminal(conn, term, handler)
				req.Reply(true, nil)
			case "pty-req":
				// p10, 6.2.  Requesting a Pseudo-Terminal, RFC4254
				if reqSize < 4 {
					log.Printf("malformed pty-req request")
					continue
				}

				termVariableSize := int(req.Payload[3])

				if reqSize < 4+termVariableSize+8 {
					log.Printf("malformed pty-req request")
					continue
				}

				w := binary.BigEndian.Uint32(req.Payload[4+termVariableSize:])
				h := binary.BigEndian.Uint32(req.Payload[4+termVariableSize+4:])

				term.SetSize(int(w), int(h))

				req.Reply(true, nil)
			case "window-change":
				// p10, 6.7.  Window Dimension Change Message, RFC4254
				if reqSize < 8 {
					log.Printf("malformed window-change request")
					continue
				}

				w := binary.BigEndian.Uint32(req.Payload)
				h := binary.BigEndian.Uint32(req.Payload[4:])

				term.SetSize(int(w), int(h))
			}
		}
	}()
}

func handleChannels(chans <-chan ssh.NewChannel, handler sshHandler) {
	for newChannel := range chans {
		go handleChannel(newChannel, handler)
	}
}

func accept(listener net.Listener, handler sshHandler, srv *ssh.ServerConfig) {
	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Printf("error accepting connection, %v", err)
			continue
		}

		sshConn, chans, reqs, err := ssh.NewServerConn(conn, srv)

		if err != nil {
			log.Printf("error accepting handshake, %v", err)
			continue
		}

		log.Printf("new ssh connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())

		go ssh.DiscardRequests(reqs)
		go handleChannels(chans, handler)
	}
}

func StartSSHServer(listener net.Listener, handler sshHandler) {
	srv := &ssh.ServerConfig{
		NoClientAuth: true,
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	if err != nil {
		log.Fatal("private key generation error: ", err)
	}

	signer, err := ssh.NewSignerFromKey(key)

	if err != nil {
		log.Fatal("key conversion error: ", err)
	}

	srv.AddHostKey(signer)

	log.Printf("starting ssh server (%s) at %s", ssh.FingerprintSHA256(signer.PublicKey()), listener.Addr())
	go accept(listener, handler, srv)
}
