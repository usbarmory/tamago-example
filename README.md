TamaGo - bare metal Go for AMD64/ARM/RISC-V processors
======================================================

tamago | https://github.com/usbarmory/tamago  

![TamaGo gopher](https://github.com/usbarmory/tamago/wiki/images/tamago.svg?sanitize=true)

Authors
=======

Andrea Barisani  
andrea@inversepath.com  

Andrej Rosano  
andrej@inversepath.com  

Introduction
============

TamaGo is a framework that enables compilation and execution of unencumbered Go
applications on bare metal AMD64/ARM/RISC-V processors.

This example Go application illustrates use of the
[tamago](https://github.com/usbarmory/tamago) package
execute bare metal Go code on the following platforms:

| Processor             | Board                                                                                                                                                                                | SoC/CPU package                                                           | Board package                                                                                    |
|-----------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------|
| NXP i.MX6ULZ/i.MX6UL  | [USB armory Mk II](https://github.com/usbarmory/usbarmory/wiki/Mk-II-Introduction)                                                                                                   | [imx6ul](https://github.com/usbarmory/tamago/tree/master/soc/nxp/imx6ul)  | [usbarmory/mk2](https://github.com/usbarmory/tamago/tree/master/board/usbarmory)                 |
| NXP i.MX6ULL/i.MX6UL  | [USB armory Mk II LAN](https://github.com/usbarmory/usbarmory/wiki/Mk-II-LAN)                                                                                                        | [imx6ul](https://github.com/usbarmory/tamago/tree/master/soc/nxp/imx6ul)  | [usbarmory/mk2](https://github.com/usbarmory/tamago/tree/master/board/usbarmory)                 |
| NXP i.MX6ULL/i.MX6ULZ | [MCIMX6ULL-EVK](https://www.nxp.com/design/development-boards/i-mx-evaluation-and-development-boards/evaluation-kit-for-the-i-mx-6ull-and-6ulz-applications-processor:MCIMX6ULL-EVK) | [imx6ul](https://github.com/usbarmory/tamago/tree/master/soc/nxp/imx6ul)  | [nxp/mx6ullevk](https://github.com/usbarmory/tamago/tree/master/board/nxp/mx6ullevk)             |
| SiFive FU540          | [QEMU sifive_u](https://www.qemu.org/docs/master/system/riscv/sifive_u.html)                                                                                                         | [fu540](https://github.com/usbarmory/tamago/tree/master/soc/sifive/fu540) | [qemu/sifive_u](https://github.com/usbarmory/tamago/tree/master/board/qemu/sifive_u)             |
| AMD/Intel 64-bit      | [QEMU microvm](https://www.qemu.org/docs/master/system/i386/microvm.html)                                                                                                            | [amd64](https://github.com/usbarmory/tamago/tree/master/amd64)            | [qemu/microvm](https://github.com/usbarmory/tamago/tree/master/board/qemu/microvm)               |
| AMD/Intel 64-bit      | [Firecracker microvm](https://firecracker-microvm.github.io)                                                                                                                         | [amd64](https://github.com/usbarmory/tamago/tree/master/amd64)            | [firecracker/microvm](https://github.com/usbarmory/tamago/tree/master/board/firecracker/microvm) |

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
  * Test BTC transaction creation and signing.
  * Test post-quantum key encapsulation (KEM).
  * Hardware accelerated encryption, hashing, key derivation (on non-emulated runs).
  * Large memory allocation.

The following network services are started:

  * SSH server on 10.0.0.1:22
  * HTTP server on 10.0.0.1:80
  * HTTPS server on 10.0.0.1:443

On the USB armory Mk II the network interface is exposed over [Ethernet over
USB](https://github.com/usbarmory/usbarmory/wiki/Host-communication) (ECM
protocol, supported on Linux and macOS hosts).

On the USB armory Mk II LAN the network interface is exposed on both USB and
physical Ethernet interfaces.

On the MCIMX6ULL-EVK the second Ethernet port is used.

On QEMU and Firecracker microVMs VirtIO networking is used.

The web servers expose the following routes:

  * `/`: a welcome message
  * `/tamago-example.log`: log output
  * `/dir`: in-memory filesystem test directory (available after `test` is issued)
  * `/debug/pprof`: Go runtime profiling data through [pprof](https://golang.org/pkg/net/http/pprof/)
  * `/debug/statsviz`: Go runtime profiling data through [statsviz](https://github.com/arl/statsviz)

The SSH server exposes a console with the following commands (i.MX6UL boards):

```
9p                                                               # start 9p remote file server
aes             <size> <sec> (soft)?                             # benchmark CAAM/DCP hardware encryption
bee             <hex region0> <hex region1>                      # BEE OTF AES memory encryption
ble                                                              # BLE serial console
date            (time in RFC339 format)?                         # show/change runtime date and time
dma             (free|used)?                                     # show allocation of default DMA region
dns             <host>                                           # resolve domain
ecdsa           <sec> (soft)?                                    # benchmark CAAM/DCP hardware signing
exit, quit                                                       # close session
halt                                                             # halt the machine
freq            (198|396|528|792|900)                            # change ARM core frequency
help                                                             # this help
huk                                                              # CAAM/DCP hardware unique key derivation
i2c             <n> <hex target> <hex addr> <size>               # I²C bus read
info                                                             # device information
kem                                                              # benchmark post-quantum KEM
ls              (path)?                                          # list directory contents
led             (white|blue) (on|off)                            # LED control
mii             <hex pa> <hex ra> (hex data)?                    # show/change eth PHY standard registers
mmd             <hex pa> <hex devad> <hex ra> (hex data)?        # show/change eth PHY extended registers
ntp             <host>                                           # change runtime date and time via NTP
otp             <bank> <word>                                    # OTP fuses display
peek            <hex offset> <size>                              # memory display (use with caution)
poke            <hex offset> <hex value>                         # memory write   (use with caution)
rand                                                             # gather 32 random bytes
reboot                                                           # reset device
rtic            (<hex start> <hex end>)?                         # start RTIC on .text and optional region
sha             <size> <sec> (soft)?                             # benchmark CAAM/DCP hardware hashing
stack                                                            # goroutine stack trace (current)
stackall                                                         # goroutine stack trace (all)
tailscale       <auth key> (verbose)?                            # start network servers on Tailscale tailnet
test                                                             # launch tests
usdhc           <n> <hex offset> <size>                          # SD/MMC card read
wormhole        (send <path>|recv <code>)                        # transfer file through magic wormhole
```

On emulated runs (e.g. `make qemu`) for `usbarmory` and `sifive_u` targets the
console is exposed directly on the terminal, otherwise networking is used.

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
make example TARGET=sifive_u
```

Available targets:

| `TARGET`    | Board            | Executing and debugging                                                                                 |
|-------------|------------------|---------------------------------------------------------------------------------------------------------|
| `sifive_u`  | QEMU sifive_u    | [sifive_u](https://github.com/usbarmory/tamago/tree/master/board/qemu/sifive_u#executing-and-debugging) |

The target has only been tested with emulated execution (e.g. `make qemu`).

Building and executing on AMD64 targets
=======================================

Build the application executables as follows:

```
make example TARGET=microvm
```

Available targets:

| `TARGET`      | Board               | Executing and debugging                                                                                                  |
|---------------|---------------------|--------------------------------------------------------------------------------------------------------------------------|
| `microvm`     | QEMU microvm        | [qemu/microvm](https://github.com/usbarmory/tamago/tree/master/board/qemu/microvm#executing-and-debugging)               |
| `firecracker` | Firecracker microvm | [firecracker/microvm](https://github.com/usbarmory/tamago/tree/master/board/firecracker/microvm#executing-and-debugging) |

Both targets are meant for paravirtualized execution, respectively with QEMU:

```
make qemu TARGET=microvm
```

Or Firecracker (shown in the example via
[firectl](https://github.com/firecracker-microvm/firectl)):

```
make example TARGET=firecracker
firectl --kernel example --root-drive /dev/null --tap-device tap0/06:00:AC:10:00:01 -c 1 -m 4096
```

In both cases networking can be configured as shown in the next section.

Emulated hardware with QEMU
===========================

QEMU supported targets can be executed under emulation as follows:

```
make qemu
```

The emulation run will either provide an interactive console or emulated
Ethernet connectivity, in the latter case tap0 should be configured as follows
(Linux example):

```
ip addr add 10.0.0.2/24 dev tap0
ip link set tap0 up
ip tuntap add dev tap0 mode tap group <your user group>
```

An emulated target can be debugged with GDB using `make qemu-gdb`, this will
make qemu waiting for a GDB connection that can be launched as follows:

```
# path should be adjusted on cross platforms (e.g. arm-none-eabi-gdb)
gdb -ex "target remote 127.0.0.1:1234" example
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
