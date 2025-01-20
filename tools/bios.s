.align 2
.include "bios.cfg"

.section .text
.globl _start

_start:
        li    t0, RT0_RISCV64_TAMAGO
        jr    t0
