// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build gcp

package cmd

import (
	"github.com/usbarmory/tamago/board/google/gcp"
	"github.com/usbarmory/tamago/kvm/clock"
)

const boardName = "gcp"

func init() {
	Terminal = gcp.UART0

	// set date and time at boot
	gcp.AMD64.SetTime(kvmclock.Now().UnixNano())
}

func date(epoch int64) {
	gcp.AMD64.SetTime(epoch)
}

func uptime() (ns int64) {
	return gcp.AMD64.GetTime() - gcp.AMD64.TimerOffset
}

func Target() (name string, freq uint32) {
	return gcp.AMD64.Name(), gcp.AMD64.Freq()
}

func HasNetwork() (usb bool, eth bool) {
	return false, true
}
