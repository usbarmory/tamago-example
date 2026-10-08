// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build microvm

package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"runtime"
	"runtime/goos"
	"strings"
	"time"
	_ "unsafe"

	"github.com/townsendmerino/goinfer/decoder"
	gitok "github.com/townsendmerino/goinfer/tokenizer"

	"github.com/usbarmory/tamago/amd64"
	"github.com/usbarmory/tamago/board/qemu/microvm"

	"github.com/usbarmory/tamago-example/shell"
)

const (
	ramStart    = 0x1_0000_0000
	ramSize     = 4 << 30 // 4 GiB
	modelURL    = "https://huggingface.co/Qwen/Qwen3-0.6B-GGUF/resolve/main/Qwen3-0.6B-Q8_0.gguf"
	quant       = "int8"
	eos         = "<|im_end|>"
	tokens      = 256
	seed        = 42
	temperature = 0.7
	probability = 0.8
)

const promptTemplate = "<|im_start|>user\n%s<|im_end|>\n<|im_start|>assistant\n<think>\n\n</think>\n\n"

//go:linkname moveHeap runtime/goos.Hwinit0
func moveHeap() {
	microvm.AMD64.ConfigurePDPT(ramStart, ramStart + ramSize, amd64.MemoryRegion)

	goos.RamStart = ramStart
	goos.RamSize = ramSize

	goos.Bloc = uintptr(goos.RamStart)
	goos.BlocMax = uintptr(goos.RamStart + goos.RamSize)
}

var (
	model     *decoder.Model
	tokenizer *gitok.Tokenizer
	eosID     int
)

func init() {
	shell.Add(shell.Cmd{
		Name:    "gpt",
		Args:    1,
		Pattern: regexp.MustCompile(`^gpt (.*)`),
		Syntax:  "<prompt>",
		Help:    "ask a local LLM model a single question",
		Fn:      gptCmd,
	})
}

func download(url string) (resp *http.Response, err error) {
	if resp, err = http.Get(url); err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		return resp, fmt.Errorf("GET %s: %s", url, resp.Status)
	}

	return
}

func loadModel() (err error) {
	var resp *http.Response
	var raw []byte

	if model != nil {
		return
	}

	log.Printf("downloading %s", modelURL)
	if resp, err = download(modelURL); err != nil {
		return
	}
	defer resp.Body.Close()

	// goinfer parses GGUF from memory, pre-size the buffer to avoid
	// io.ReadAll growth transiently doubling the allocation
	if n := resp.ContentLength; n > 0 {
		raw = make([]byte, n)
		_, err = io.ReadFull(resp.Body, raw)
	} else {
		raw, err = io.ReadAll(resp.Body)
	}

	if err != nil {
		return
	}

	log.Printf("loading model")
	m, err := decoder.LoadGGUFBytes(raw, decoder.Options{Backend: "cpu", Quant: quant})

	if err != nil {
		return
	}

	log.Printf("parsing tokenizer")
	tok, err := gitok.LoadGGUFBytes(raw)

	if err != nil {
		m.Close()
		return
	}

	id, ok := tok.TokenID(eos)

	if !ok {
		m.Close()
		return fmt.Errorf("tokenizer: missing %q token", eos)
	}

	model, tokenizer, eosID = m, tok, id

	return
}

func ask(question string) (answer string, err error) {
	prompt, err := tokenizer.Encode(fmt.Sprintf(promptTemplate, question), false)

	if err != nil {
		return
	}

	start := time.Now()

	ch, gen := model.Generate(context.Background(), prompt, tokens, decoder.SamplingParams{
		Temperature: temperature,
		TopP:        probability,
		Seed:        seed,
		StopIDs:     []int{eosID},
	})

	var out []int

	for id := range ch {
		out = append(out, id)
	}

	if err = gen.Err(); err != nil {
		return
	}

	elapsed := time.Since(start)

	text, err := tokenizer.Decode(out)

	if err != nil {
		return
	}

	answer = fmt.Sprintf("%d tok in %v (%.2f tok/s)\n", len(out), elapsed, float64(len(out))/elapsed.Seconds())
	answer += strings.TrimSpace(text)

	return
}

func gptCmd(console *shell.Interface, arg []string) (res string, err error) {
	defer runtime.GC()

	if err = loadModel(); err != nil {
		return
	}

	cpuidleCmd(nil, []string{"off"})
	defer cpuidleCmd(nil, []string{"on"})

	return ask(arg[0])
}
