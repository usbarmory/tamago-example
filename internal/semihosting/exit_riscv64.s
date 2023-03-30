// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func sys_exit()
TEXT ·sys_exit(SB), $0
	MOV	$0x18,    A0
	MOV	$0x20026, A1

	SLLI	$0x1f, X0, X0
	EBREAK
	SRAI	$0x07, X0, X0
	RET
