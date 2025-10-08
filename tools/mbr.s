; Copyright (c) The TamaGo Authors. All Rights Reserved.
;
; Use of this source code is governed by the license
; that can be found in the LICENSE file.

; Implement a simple Master Boot Record that set up 32-bit protected mode,
; reads TamaGo image from disk and jumps to it.
; Reference: https://wiki.osdev.org/MBR_(x86)

base equ 0x7c00
buffer equ 0x10000 ; 0x10000 - 0x1ffff: 64kB readsectors buffer

%macro ENTER_REAL 0
	bits 32
	mov [stack.prot], esp
	jmp gdt32.code16:%%prot16

%%prot16:
	bits 16
	mov eax, cr0
	and eax, ~1
	mov cr0, eax
	jmp 0:%%real

%%real:
	xor ax, ax
	mov ds, ax
	mov ss, ax
	mov sp, [stack.real]
%endmacro

%macro ENTER_PROTECTED 0
	bits 16
	mov eax, cr0
	or eax, 1
	mov cr0, eax
	jmp gdt32.code32:%%protected

%%protected:
	bits 32
	mov eax, gdt32.data32
	mov ds, eax
	mov ss, eax
	mov esp, [stack.prot]
%endmacro

init:
	bits 16
	org base

	xor ax, ax
	mov ds, ax
	mov ss, ax

	; disable 8259 PIC
	mov al, 0xff
	out 0xa1, al
	out 0x21, al

	cli
	mov eax, base
	mov [stack.real], ax
	sub eax, 0x100
	mov [stack.prot], eax

	lgdt [gdt32.desc]
	ENTER_PROTECTED

read_tamago:
	bits 32
	ENTER_REAL
	call readsectors
	ENTER_PROTECTED
	call memcpy32
	;increment lba
	mov eax, [dap.lba]
	add eax, 64
	mov [dap.lba], eax
	mov eax, [tamago.sector_size]
	sub eax, 64
	mov [tamago.sector_size], eax
	cmp eax, 0
	jg read_tamago
	jmp [tamago.entry]

; 32-bit GDT
	align 4
gdt32:
	dw 0,0,0,0
.code32 equ $ - gdt32
	dw 0xffff,0,0x9a00,0xcf
.data32 equ $ - gdt32
	dw 0xffff,0,0x9200,0xcf
.code16 equ $ - gdt32
	dw 0xffff,0,0x9a00,0x00
.data16 equ $ - gdt32
	dw 0xffff,0,0x9200,0x00
.desc:
	dw $ - gdt32 -1
	dd gdt32

; real and protected stack pointer saving location
stack:
	.real	dw 0
	.prot	dd 0

; disk address packet structure
dap:
	db 0x10
	db 0
	.sector_count	dw 64			; 32 kB = 0x8000
	.offset		dw 0
	.segment	dw (buffer >> 4)	; readsector buffer: 0x10000-0x1ffff
	.lba		dq 200

; TAMAGO_* parameters should be provided externally
tamago:
	.offset		dq TAMAGO_OFFSET	; tamago image start sector
	.sector_size	dq TAMAGO_SIZE/512	; tamago size in sectors
	.start		dq TAMAGO_START		; tamago binary destination address
	.entry		dq TAMAGO_ENTRY		; tamago entry point

; read disk sectors using BIOS interrupt call (INT 13h)
readsectors:
	bits 16
	mov si, dap
	mov ah, 0x42	; Extended Read Sectors From Drive
	mov dl, 0x80	; First Hard Drive
	int 0x13
	ret

memcpy32:
	bits 32
	mov ebx, 0
	mov ecx, [tamago.start]
	mov edx, buffer
copy:	mov eax, [edx]
	mov [ecx], eax
	add ecx, 4
	add edx, 4
	add ebx, 4
	cmp ebx, 0x8000		; copy 32kB chunks
	jne copy
	mov [tamago.start], ecx	; save pointer for next chunk
	ret

times (446-$+$$) db 0	; ensure output binary is 512 bytes
times 64 db 0		; partition table
dw 0xAA55		; mbr signature
