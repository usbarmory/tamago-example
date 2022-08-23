TamaGo - bare metal Go for ARM/RISCV-V SoCs - example application
=================================================================

tamago | https://github.com/usbarmory/tamago  

Copyright (c) WithSecure Corporation  
https://foundry.withsecure.com

![TamaGo gopher](https://github.com/usbarmory/tamago/wiki/images/tamago.svg?sanitize=true)

Authors
=======

Andrea Barisani  
andrea.barisani@withsecure.com | andrea@inversepath.com  

Andrej Rosano  
andrej.rosano@withsecure.com   | andrej@inversepath.com  

Introduction
============

TamaGo is a framework that enables compilation and execution of unencumbered Go
applications on bare metal ARM/RISC-V System-on-Chip (SoC) components.

This example Go application illustrates use of the
[tamago](https://github.com/usbarmory/tamago) package
execute bare metal Go code on the following platforms:

| SoC          | Board                                                                                                                                                                                | SoC package                                                               | Board package                                                                         |
|--------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------|---------------------------------------------------------------------------------------|
| NXP i.MX6ULZ | [USB armory Mk II](https://github.com/usbarmory/usbarmory/wiki)                                                                                                                      | [imx6ul](https://github.com/usbarmory/tamago/tree/master/soc/nxp/imx6ul)  | [usbarmory/mk2](https://github.com/usbarmory/tamago/tree/master/board/usbarmory)      |
| NXP i.MX6ULL | [MCIMX6ULL-EVK](https://www.nxp.com/design/development-boards/i-mx-evaluation-and-development-boards/evaluation-kit-for-the-i-mx-6ull-and-6ulz-applications-processor:MCIMX6ULL-EVK) | [imx6ul](https://github.com/usbarmory/tamago/tree/master/soc/nxp/imx6ul)  | [nxp/mx6ullevk](https://github.com/usbarmory/tamago/tree/master/board/nxp/mx6ullevk)  |
| SiFive FU540 | [QEMU sifive_u](https://www.qemu.org/docs/master/system/riscv/sifive_u.html)                                                                                                         | [fu540](https://github.com/usbarmory/tamago/tree/master/soc/sifive/fu540) | [qemu/sifive_u](https://github.com/usbarmory/tamago/tree/master/board/qemu/sifive_u)  |

Documentation
=============

For more information about TamaGo see its
[repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki).

Operation
=========

![Example screenshot](https://github.com/usbarmory/tamago/wiki/images/ssh.png)

The example application performs a variety of simple test procedures, each in
its separate goroutine, to demonstrate bare metal execution of Go standard and
external libraries:

  * Directory and file write/read from an in-memory filesystem.
  * SD/MMC card detection and read (only on non-emulated runs).
  * Timer operation.
  * Sleep operation.
  * Random bytes collection (gathered from SoC TRNG on non-emulated runs).
  * ECDSA signing and verification.
  * Test BTC transaction creation and signing.
  * Test post-quantum key encapsulation (KEM).
  * Hardware key derivation (only on non-emulated runs).
  * Large memory allocation.

On non-emulated hardware the following network services are started on
[Ethernet over USB](https://github.com/usbarmory/usbarmory/wiki/Host-communication) (ECM
protocol, supported on Linux and macOS hosts).

  * SSH server on 10.0.0.1:22
  * HTTP server on 10.0.0.1:80
  * HTTPS server on 10.0.0.1:443

The web servers expose the following routes:

  * `/`: a welcome message
  * `/tamago-example.log`: log output
  * `/dir`: in-memory filesystem test directory (available after `test` is issued)
  * `/debug/pprof`: Go runtime profiling data through [pprof](https://golang.org/pkg/net/http/pprof/)
  * `/debug/statsviz`: Go runtime profiling data through [statsviz](https://github.com/arl/statsviz)

The SSH server exposes a console with the following commands:

```

ble                                                      # BLE serial console
date            (time in RFC339 format)?                 # show/change runtime date and time
dcp             <size> <sec>                             # benchmark hardware encryption
dns             <fqdn>                                   # resolve domain (requires routing)
exit, quit                                               # close session
help                                                     # this help
i2c             <n> <hex target> <hex addr> <size>       # I²C bus read
info                                                     # device information
kem                                                      # benchmark post-quantum KEM
led             (white|blue) (on|off)                    # LED control
mmc             <n> <hex offset> <size>                  # MMC/SD card read
otp             <bank> <word>                            # OTP fuses display
peek            <hex offset> <size>                      # memory display (use with caution)
poke            <hex offset> <hex value>                 # memory write   (use with caution)
rand                                                     # gather 32 random bytes
reboot                                                   # reset device
stack                                                    # stack trace of current goroutine
stackall                                                 # stack trace of all goroutines
test                                                     # launch tests
```

On emulated runs (e.g. `make qemu`) the console is exposed directly on the
terminal.

Building the compiler
=====================

Build the [TamaGo compiler](https://github.com/usbarmory/tamago-go)
(or use the [latest binary release](https://github.com/usbarmory/tamago-go/releases/latest)):

```
wget https://github.com/usbarmory/tamago-go/archive/refs/tags/latest.zip
unzip latest.zip
cd tamago-go-latest/src && ./all.bash
cd ../bin && export TAMAGO=`pwd`/go
```

Building and executing on ARM targets
=====================================

Build the application executables as follows:

```
make imx TARGET=usbarmory
```

The following targets are available:

| `TARGET`    | Board            | Executing and debugging                                                                                  |
|-------------|------------------|----------------------------------------------------------------------------------------------------------|
| `usbarmory` | USB armory Mk II | [usbarmory](https://github.com/usbarmory/tamago/tree/master/board/usbarmory#executing-and-debugging)     |
| `mx6ullevk` | MCIMX6ULL-EVK    | [mx6ullevk](https://github.com/usbarmory/tamago/tree/master/board/nxp/mx6ullevk#executing-and-debugging) |

The targets support native (see relevant documentation links in the table above)
as well as emulated execution (e.g. `make qemu`).

Building and executing on RISC-V targets
========================================

Build the application executables as follows:

```
make TARGET=sifive_u
```

Available targets:

| `TARGET`    | Board            | Executing and debugging                                                                                  |
|-------------|------------------|----------------------------------------------------------------------------------------------------------|
| `sifive_u`  | QEMU sifive_u    | [sifive_u](https://github.com/usbarmory/tamago/tree/master/board/qemu/sifive_u#executing-and-debugging)  |

The target has only been tested with emulated execution (e.g. `make qemu`)

Emulated hardware with QEMU
===========================

All targets can be executed under emulation as follows:

```
make qemu
```

An emulated target can be debugged with GDB using `make qemu-gdb`, this will
make qemu waiting for a GDB connection that can be launched as follows:

```
# ARM targets
arm-none-eabi-gdb -ex "target remote 127.0.0.1:1234" example

# RISC-V targets
riscv64-elf-gdb -ex "target remote 127.0.0.1:1234" example
```

Breakpoints can be set in the usual way:

```
b ecdsa.Verify
continue
```

License
=======

tamago | https://github.com/usbarmory/tamago  
Copyright (c) WithSecure Corporation

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/tamago-example/blob/master/LICENSE) file.

The TamaGo logo is adapted from the Go gopher designed by Renee French and
licensed under the Creative Commons 3.0 Attributions license. Go Gopher vector
illustration by Hugo Arganda.
