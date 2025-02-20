// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !(firecracker || microvm || mx6ullevk || usbarmory)

package network

import (
	"log"
)

func Init(_ any, _ bool, _ bool, _ interface{}) (_ any) {
	log.Fatal("unsupported")
	return
}
