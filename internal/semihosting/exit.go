// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build semihosting && !amd64

package semihosting

// defined in exit_$GOARCH.s
func sys_exit()

func Exit() {
	sys_exit()
}
