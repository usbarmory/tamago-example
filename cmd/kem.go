// The following functions are adapted from:
//   https://github.com/FiloSottile/mlkem768/blob/main/xwing/xwing_test.go
//
// See LICENSE at:
//   https://github.com/FiloSottile/mlkem768/blob/main/LICENSE

package cmd

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"time"

	"filippo.io/mlkem768/xwing"
	"golang.org/x/term"

	"github.com/usbarmory/tamago-example/shell"
)

func init() {
	shell.Add(shell.Cmd{
		Name: "kem",
		Help: "benchmark post-quantum KEM",
		Fn:   kemCmd,
	})
}

func kemCmd(_ *shell.Interface, _ *term.Terminal, arg []string) (res string, err error) {
	return "", xwingRoundTrip(log.Default())
}

func kemTest() (tag string, res string) {
	tag = "post-quantum KEM"

	b := &strings.Builder{}
	log := log.New(b, "", 0)

	xwingRoundTrip(log)

	return tag, b.String()
}

func xwingRoundTrip(log *log.Logger) (err error) {
	start := time.Now()

	dk, err := xwing.GenerateKey()
	if err != nil {
		return
	}
	c, Ke, err := xwing.Encapsulate(dk.EncapsulationKey())
	if err != nil {
		return
	}
	Kd, err := xwing.Decapsulate(dk, c)
	if err != nil {
		return
	}
	if !bytes.Equal(Ke, Kd) {
		return errors.New("Ke != Kd")
	}

	dk1, err := xwing.GenerateKey()
	if err != nil {
		return
	}
	if bytes.Equal(dk.EncapsulationKey(), dk1.EncapsulationKey()) {
		return errors.New("ek == ek1")
	}
	if bytes.Equal(dk.Bytes(), dk1.Bytes()) {
		return errors.New("dk == dk1")
	}

	dk2, err := xwing.NewKeyFromSeed(dk.Bytes())
	if err != nil {
		return
	}
	if !bytes.Equal(dk.Bytes(), dk2.Bytes()) {
		return errors.New("dk != dk2")
	}

	c1, Ke1, err := xwing.Encapsulate(dk.EncapsulationKey())
	if err != nil {
		return
	}
	if bytes.Equal(c, c1) {
		return errors.New("c == c1")
	}
	if bytes.Equal(Ke, Ke1) {
		return errors.New("Ke == Ke1")
	}

	log.Printf("xwing-kem %x (%s)", dk.Bytes(), time.Since(start))

	return
}
