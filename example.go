// https://github.com/f-secure-foundry/tamago-example
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Basic test example for tamago/arm running on supported i.MX6 targets.

package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"math"
	"math/big"
	mathrand "math/rand"
	"os"
	"runtime"
	"time"

	"github.com/f-secure-foundry/tamago/soc/imx6"
)

var Build string
var Revision string
var banner string

var exit chan bool

var logFile *os.File

func init() {
	log.SetFlags(0)

	banner = fmt.Sprintf("%s/%s (%s) • %s %s",
		runtime.GOOS, runtime.GOARCH, runtime.Version(),
		Revision, Build)

	model := imx6.Model()

	if !imx6.Native {
		banner += fmt.Sprintf(" • %s %d MHz (emulated)", model, imx6.ARMFreq()/1000000)
		return
	}

	if err := imx6.SetARMFreq(900); err != nil {
		log.Printf("WARNING: error setting ARM frequency: %v", err)
	}

	banner += fmt.Sprintf(" • %s %d MHz", model, imx6.ARMFreq()/1000000)
}

func test(init bool) {
	start := time.Now()
	exit = make(chan bool)
	n := 0

	log.Println("-- begin tests -------------------------------------------------------")

	n += 1
	go func() {
		log.Println("-- fs ----------------------------------------------------------------")
		TestDev()
		TestFile()

		exit <- true
	}()

	sleep := 100 * time.Millisecond

	n += 1
	go func() {
		log.Println("-- timer -------------------------------------------------------------")

		t := time.NewTimer(sleep)
		log.Printf("waking up timer after %v", sleep)

		start := time.Now()

		for now := range t.C {
			log.Printf("woke up at %d (%v)", now.Nanosecond(), now.Sub(start))
			break
		}

		exit <- true
	}()

	n += 1
	go func() {
		log.Println("-- sleep -------------------------------------------------------------")

		log.Printf("sleeping %s", sleep)
		start := time.Now()
		time.Sleep(sleep)
		log.Printf("slept %s (%v)", sleep, time.Since(start))

		exit <- true
	}()

	n += 1
	go func() {
		log.Println("-- rng ---------------------------------------------------------------")

		size := 32

		for i := 0; i < 10; i++ {
			rng := make([]byte, size)
			rand.Read(rng)
			log.Printf("%x", rng)
		}

		count := 1000
		start := time.Now()

		for i := 0; i < count; i++ {
			rng := make([]byte, size)
			rand.Read(rng)
		}

		log.Printf("retrieved %d random bytes in %s", size*count, time.Since(start))

		seed, _ := rand.Int(rand.Reader, big.NewInt(int64(math.MaxInt64)))
		mathrand.Seed(seed.Int64())

		exit <- true
	}()

	n += 1
	go func() {
		log.Println("-- ecdsa -------------------------------------------------------------")
		TestSignAndVerify()
		exit <- true
	}()

	n += 1
	go func() {
		log.Println("-- btc ---------------------------------------------------------------")

		ExamplePayToAddrScript()
		ExampleExtractPkScriptAddrs()
		ExampleSignTxOutput()

		exit <- true
	}()

	if imx6.Native && imx6.Family == imx6.IMX6ULL {
		n += 1
		go func() {
			log.Println("-- i.mx6 dcp ---------------------------------------------------------")
			TestDCP()
			exit <- true
		}()
	}

	log.Printf("launched %d test goroutines", n)

	for i := 1; i <= n; i++ {
		<-exit
	}

	log.Printf("----------------------------------------------------------------------")
	log.Printf("completed %d goroutines (%s)", n, time.Since(start))

	runs := 9
	chunksMax := 50
	chunks := mathrand.Intn(chunksMax) + 1
	fillSize := 160 * 1024 * 1024
	chunkSize := fillSize / chunks

	log.Printf("-- memory allocation (%d runs) ----------------------------------------", runs)
	testAlloc(runs, chunks, chunkSize)

	if imx6.Native {
		size := 10 * 1024 * 1024
		readSize := 0x7fff

		if init {
			// We can use the entire iRAM before USB activation,
			// accounting for required dTD alignment which takes
			// additional space.
			readSize = 0x20000 - 4096
		}

		log.Println("-- memory cards -------------------------------------------------------")

		for _, card := range cards {
			TestUSDHC(card, size, readSize)
		}
	}
}

func main() {
	start := time.Now()

	logFile, _ = os.OpenFile("/tamago-example.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	log.Println(banner)

	test(true)

	if imx6.Native && (imx6.Family == imx6.IMX6UL || imx6.Family == imx6.IMX6ULL) {
		log.Println("-- i.mx6 usb ---------------------------------------------------------")
		startNetworking()
	}

	log.Printf("Goodbye from tamago/arm (%s)", time.Since(start))
}
