riscv64-linux-gnu-gcc -march=rv64g -mabi=lp64 -static -mcmodel=medany -fvisibility=hidden -nostdlib -nostartfiles -Tbios.ld bios.s -o bios.bin
