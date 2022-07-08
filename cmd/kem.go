// The following functions are adapted from:
//   https://github.com/cloudflare/circl/blob/v1.2.0/kem/kyber/kat_test.go
//
// See LICENSE at:
//   https://github.com/cloudflare/circl/blob/v1.2.0/LICENSE

// The build restriction is required as golang.org/x/sys/cpu is missing riscv64
// stubs (see https://github.com/golang/go/issues/53698).
//
//go:build mx6ullevk || usbarmory
// +build mx6ullevk usbarmory

package cmd

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"log"
	"time"

	"github.com/cloudflare/circl/kem/schemes"
	"golang.org/x/term"

	"github.com/usbarmory/tamago-example/internal/nist"
)

func init() {
	Add(Cmd{
		Name: "kem",
		Args: 0,
		Help: "benchmark post-quantum KEM",
		Fn:   kemCmd,
	})
}

func kemCmd(_ *term.Terminal, arg []string) (res string, err error) {
	kats := []struct {
		name string
		want string
	}{
		// Computed from reference implementation
		{"Kyber1024", "89248f2f33f7f4f7051729111f3049c409a933ec904aedadf035f30fa5646cd5"},
		{"Kyber768", "a1e122cad3c24bc51622e4c242d8b8acbcd3f618fee4220400605ca8f9ea02c2"},
		{"Kyber512", "e9c2bd37133fcb40772f81559f14b1f58dccd1c816701be9ba6214d43baf4547"},
	}

	for _, kat := range kats {
		testPQCgenKATKem(kat.name, kat.want)
	}

	return
}

func kyberTest() {
	msg("post-quantum KEM")
	testPQCgenKATKem("Kyber512", "e9c2bd37133fcb40772f81559f14b1f58dccd1c816701be9ba6214d43baf4547")
}

func testPQCgenKATKem(name, expected string) {
	scheme := schemes.ByName(name)
	if scheme == nil {
		log.Fatal("invalid scheme")
	}

	start := time.Now()

	var seed [48]byte
	kseed := make([]byte, scheme.SeedSize())
	eseed := make([]byte, scheme.EncapsulationSeedSize())
	for i := 0; i < 48; i++ {
		seed[i] = byte(i)
	}
	f := sha256.New()
	g := nist.NewDRBG(&seed)
	fmt.Fprintf(f, "# %s\n\n", name)
	for i := 0; i < 100; i++ {
		g.Fill(seed[:])
		fmt.Fprintf(f, "count = %d\n", i)
		fmt.Fprintf(f, "seed = %X\n", seed)
		g2 := nist.NewDRBG(&seed)

		// This is not equivalent to g2.Fill(kseed[:]).  As the reference
		// implementation calls randombytes twice generating the keypair,
		// we have to do that as well.
		g2.Fill(kseed[:32])
		g2.Fill(kseed[32:])

		g2.Fill(eseed)
		pk, sk := scheme.DeriveKeyPair(kseed)
		ppk, _ := pk.MarshalBinary()
		psk, _ := sk.MarshalBinary()
		ct, ss, _ := scheme.EncapsulateDeterministically(pk, eseed)
		ss2, _ := scheme.Decapsulate(sk, ct)
		if !bytes.Equal(ss, ss2) {
			log.Fatal("decapsulation mismatch")
		}
		fmt.Fprintf(f, "pk = %X\n", ppk)
		fmt.Fprintf(f, "sk = %X\n", psk)
		fmt.Fprintf(f, "ct = %X\n", ct)
		fmt.Fprintf(f, "ss = %X\n\n", ss)
	}

	if fmt.Sprintf("%x", f.Sum(nil)) != expected {
		log.Fatal("hash mismatch")
	}

	log.Printf("%-9s %x (%s)", name, f.Sum(nil), time.Since(start))
}
