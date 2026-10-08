// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build microvm

package cmd

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"runtime"
	"runtime/goos"
	"strings"
	"time"
	_ "unsafe"

	_ "github.com/jxsl13/goai/backend/cpu"
	_ "github.com/jxsl13/goai/backend/ref"
	"github.com/jxsl13/goai/format/gguf"
	"github.com/jxsl13/goai/nlp"

	"github.com/usbarmory/tamago/amd64"
	"github.com/usbarmory/tamago/board/qemu/microvm"

	"github.com/usbarmory/tamago-example/shell"
)

const (
	ramStart    = 0x1_0000_0000
	ramSize     = 4 << 30 // 4 GiB
	modelURL    = "https://huggingface.co/Qwen/Qwen3-0.6B-GGUF/resolve/main/Qwen3-0.6B-Q8_0.gguf"
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
	model     *nlp.QuantLlama
	tokenizer *nlp.BPETokenizer
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

	if model != nil {
		return
	}

	log.Printf("downloading %s", modelURL)
	if resp, err = download(modelURL); err != nil {
		return
	}
	defer resp.Body.Close()

	modelBuf := bufio.NewReaderSize(resp.Body, 1<<20)

	log.Printf("loading model")
	rf, err := gguf.ReadRaw(modelBuf)

	if err != nil {
		return
	}

	log.Printf("parsing model")
	m, err := nlp.QuantQwen3FromGGUF(rf.Metadata, rf.Tensors)
	if err != nil {
		return
	}

	log.Printf("parsing tokenizer")
	tok, err := nlp.BPEFromGGUF(rf.Metadata)
	if err != nil {
		return
	}

	ids := tok.EncodeSpecial(eos)
	if len(ids) != 1 {
		return fmt.Errorf("tokenizer: %q is not a single special token (got %v)", eos, ids)
	}

	model, tokenizer, eosID = m, tok, ids[0]

	return
}

func ask(question string) (answer string, err error) {
	prompt := tokenizer.EncodeSpecial(fmt.Sprintf(promptTemplate, question))

	start := time.Now()

	out, err := model.Generate(
		prompt,
		tokens,
		nlp.NewSampler(seed, nlp.WithTemperature(temperature), nlp.WithTopP(probability)),
		nlp.WithEOS(eosID),
	)

	if err != nil {
		return
	}

	elapsed := time.Since(start)

	gen := out[len(prompt):]

	if n := len(gen); n > 0 && gen[n-1] == eosID {
		gen = gen[:n-1]
	}

	answer = fmt.Sprintf("%d tok in %v (%.2f tok/s)\n", len(gen), elapsed, float64(len(gen))/elapsed.Seconds())
	answer += strings.TrimSpace(tokenizer.Decode(gen))

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
