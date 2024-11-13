// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func sys_exit()
TEXT ·sys_exit(SB), $0
	MOVL	$0xf4, DX
	MOVL	$5, AX
	WORD	$0xef	// out dx, eax
