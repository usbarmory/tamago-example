// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package cmd

import (
	"bytes"
	"crypto/mlkem"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/usbarmory/tamago-example/shell"
)

func init() {
	shell.Add(shell.Cmd{
		Name: "kem",
		Help: "benchmark post-quantum KEM",
		Fn:   kemCmd,
	})
}

func kemCmd(_ *shell.Interface, arg []string) (res string, err error) {
	return "", kemRoundTrip(log.Default())
}

func kemTest() (tag string, res string) {
	tag = "post-quantum KEM"

	b := &strings.Builder{}
	l := log.New(b, "", 0)
	l.SetPrefix(log.Prefix())

	kemRoundTrip(l)

	return tag, b.String()
}

func kemRoundTrip(log *log.Logger) (err error) {
	start := time.Now()

	dk, err := mlkem.GenerateKey768()

	if err != nil {
		return
	}

	ek := dk.EncapsulationKey()
	Ke, c := ek.Encapsulate()
	Kd, err := dk.Decapsulate(c)

	if err != nil {
		return
	}

	if !bytes.Equal(Ke, Kd) {
		return errors.New("Ke != Kd")
	}

	dk1, err := mlkem.GenerateKey768()

	if err != nil {
		return
	}

	if bytes.Equal(ek.Bytes(), dk1.EncapsulationKey().Bytes()) {
		return errors.New("ek == ek1")
	}

	if bytes.Equal(dk.Bytes(), dk1.Bytes()) {
		return errors.New("dk == dk1")
	}

	dk2, err := mlkem.NewDecapsulationKey768(dk.Bytes())

	if err != nil {
		return
	}

	if !bytes.Equal(dk.Bytes(), dk2.Bytes()) {
		return errors.New("dk != dk2")
	}

	log.Printf("mlkem %x (%s)", dk.Bytes(), time.Since(start))

	return
}
