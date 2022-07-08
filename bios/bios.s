.align 2
.include "cfg.inc"

.section .text
.globl _start

_start:
        li    t0, RT0_RISCV64_TAMAGO
        jr    t0
