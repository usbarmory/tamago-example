// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func sys_exit()
TEXT ·sys_exit(SB), $0
	MOVW	$0x18,    R0	// TARGET_SYS_EXIT
	MOVW	$0x20026, R1	// ADP_Stopped_ApplicationExit
	WORD	$0xef123456	// svc 0x00123456
	RET
