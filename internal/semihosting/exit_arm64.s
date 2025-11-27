// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func sys_exit()
TEXT ·sys_exit(SB),$0
	// ADP_Stopped_ApplicationExit
	MOVD	$0x20026, R0
	MOVD	R0, (0*8)(RSP)

	// exit code
	MOVD	$0, (1*8)(RSP)

	MOVD	$0x18, R0	// TARGET_SYS_EXIT
	MOVD	RSP, R1
	HLT	$0xf000

	RET
