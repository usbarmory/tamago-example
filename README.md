TamaGo - bare metal Go
======================

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
applications on bare metal processors (AMD64, ARM, ARM64, RISCV64).

This example Go application illustrates use of the
[tamago](https://github.com/usbarmory/tamago) package
execute bare metal Go code on the following platforms:

| Processor             | Platform                                                                                                                                                                             | SoC/CPU package                                                           | Support package                                                                                  |
|-----------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------|
| AMD/Intel 64-bit      | [Cloud Hypervisor](https://www.cloudhypervisor.org)                                                                                                                                  | [amd64](https://github.com/usbarmory/tamago/tree/master/amd64)            | [cloud_hypervisor/vm](https://github.com/usbarmory/tamago/tree/master/board/cloud_hypervisor/vm) |
| AMD/Intel 64-bit      | [QEMU microvm](https://www.qemu.org/docs/master/system/i386/microvm.html)                                                                                                            | [amd64](https://github.com/usbarmory/tamago/tree/master/amd64)            | [qemu/microvm](https://github.com/usbarmory/tamago/tree/master/board/qemu/microvm)               |
| AMD/Intel 64-bit      | [Firecracker microvm](https://firecracker-microvm.github.io)                                                                                                                         | [amd64](https://github.com/usbarmory/tamago/tree/master/amd64)            | [firecracker/microvm](https://github.com/usbarmory/tamago/tree/master/board/firecracker/microvm) |
| AMD/Intel 64-bit      | [Google Compute Engine](https://cloud.google.com/products/compute)                                                                                                                   | [amd64](https://github.com/usbarmory/tamago/tree/master/amd64)            | [google/gcp](https://github.com/usbarmory/tamago/tree/master/board/google/gcp)                   |
| NXP i.MX6ULZ/i.MX6UL  | [USB armory Mk II](https://github.com/usbarmory/usbarmory/wiki/Mk-II-Introduction)                                                                                                   | [imx6ul](https://github.com/usbarmory/tamago/tree/master/soc/nxp/imx6ul)  | [usbarmory/mk2](https://github.com/usbarmory/tamago/tree/master/board/usbarmory)                 |
| NXP i.MX6ULL/i.MX6UL  | [USB armory Mk II LAN](https://github.com/usbarmory/usbarmory/wiki/Mk-II-LAN)                                                                                                        | [imx6ul](https://github.com/usbarmory/tamago/tree/master/soc/nxp/imx6ul)  | [usbarmory/mk2](https://github.com/usbarmory/tamago/tree/master/board/usbarmory)                 |
| NXP i.MX6ULL/i.MX6ULZ | [MCIMX6ULL-EVK](https://www.nxp.com/design/development-boards/i-mx-evaluation-and-development-boards/evaluation-kit-for-the-i-mx-6ull-and-6ulz-applications-processor:MCIMX6ULL-EVK) | [imx6ul](https://github.com/usbarmory/tamago/tree/master/soc/nxp/imx6ul)  | [nxp/mx6ullevk](https://github.com/usbarmory/tamago/tree/master/board/nxp/mx6ullevk)             |
| NXP i.MX8M Plus       | [8MPLUSLPD4-EVK](https://www.nxp.com/design/design-center/development-boards-and-designs/8MPLUSLPD4-EVK)                                                                             | [imx8mp](https://github.com/usbarmory/tamago/tree/master/soc/nxp/imx8mp)  | [imx8mpevk](https://github.com/usbarmory/tamago/tree/master/board/nxp/imx8mpevk)                 |
| SiFive FU540          | [QEMU sifive_u](https://www.qemu.org/docs/master/system/riscv/sifive_u.html)                                                                                                         | [fu540](https://github.com/usbarmory/tamago/tree/master/soc/sifive/fu540) | [qemu/sifive_u](https://github.com/usbarmory/tamago/tree/master/board/qemu/sifive_u)             |

> [!NOTE]
> TamaGo also supports [UEFI](https://uefi.org/), see [go-boot](https://github.com/usbarmory/go-boot/)

Documentation
=============

[![Go Reference](https://pkg.go.dev/badge/github.com/usbarmory/tamago.svg)](https://pkg.go.dev/github.com/usbarmory/tamago)

For more information about TamaGo see its
[repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki).

Operation
=========

The example application performs a variety of simple test procedures, each in
its separate goroutine, to demonstrate bare metal execution of Go standard and
external libraries:

  * Directory and file write/read from an in-memory filesystem.
  * SD/MMC card detection and read (only on non-emulated runs, if available).
  * Timer operation.
  * Sleep operation.
  * Random bytes collection (gathered from SoC TRNG on non-emulated runs).
  * Test BTC transaction creation and signing.
  * Test post-quantum key encapsulation (KEM).
  * Hardware accelerated encryption (on non-emulated runs, if available).
  * Large memory allocation.

The following network services are started:

  * SSH server on 10.0.0.1:22
  * HTTP server on 10.0.0.1:80
  * HTTPS server on 10.0.0.1:443

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
build                                                            # build information
cat             <path>                                           # show file contents
cpuidle         (on|off)                                         # CPU idle time management control
date            (<time in RFC339 format>)?                       # show/change runtime date and time
dma             (free|used)?                                     # show allocation of default DMA region
dns             <host>                                           # resolve domain
ecdsa           <sec> (soft)?                                    # benchmark CAAM/DCP hardware signing
exit, quit                                                       # close session
hab             <srk table hash>                                 # HAB activation (use with extreme caution)
halt                                                             # halt the machine
freq            (198|396|528|792|900)                            # change ARM core frequency
help                                                             # this help
huk                                                              # CAAM/DCP hardware unique key derivation
i2c             <n> <hex target> <hex addr> <size>               # I²C bus read
info                                                             # device information
kem                                                              # benchmark post-quantum KEM
led             (white|blue) (on|off)                            # LED control
ls              (<path>)?                                        # list directory contents
mii             <hex pa> <hex ra> (hex data)?                    # show/change eth PHY standard registers
mmd             <hex pa> <hex devad> <hex ra> (hex data)?        # show/change eth PHY extended registers
ntp             <host>                                           # change runtime date and time via NTP
otp             <bank> <word>                                    # OTP fuses display
peek            <hex addr> <size>                                # memory display (use with caution)
poke            <hex addr> <hex value>                           # memory write   (use with caution)
rand                                                             # gather 32 random bytes
reboot                                                           # reset device
rtic            (<hex start> <hex end>)?                         # start RTIC on .text and optional region
sha             <size> <sec> (soft)?                             # benchmark CAAM/DCP hardware hashing
stack                                                            # goroutine stack trace (current)
stackall                                                         # goroutine stack trace (all)
tailscale       <auth key> (verbose)?                            # start network servers on Tailscale tailnet
test                                                             # launch tests
uptime                                                           # show system running time
usdhc           <n> <hex addr> <size>                            # SD/MMC card read
wormhole        (send <path>|recv <code>)                        # transfer file through magic wormhole
```

Building the compiler
=====================

The [TamaGo compiler](https://github.com/usbarmory/tamago-go) is automatically
downloaded and compiled as a `go tool` by the `Makefile`.

Alternatively the `TAMAGO` environment variable can overridden to use the
[latest binary release](https://github.com/usbarmory/tamago-go/releases/latest):

```sh
wget https://github.com/usbarmory/tamago-go/archive/refs/tags/latest.zip
unzip latest.zip
cd tamago-go-latest/src && ./all.bash
cd ../bin && export TAMAGO=`pwd`/go
```

Building and executing on AMD64 targets
=======================================

| `TARGET`           | Platform                                                                  | Executing and debugging                                                                                                  | Interface          |
|--------------------|---------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------|--------------------|
| `cloud_hypervisor` | [Cloud Hypervisor](https://www.cloudhypervisor.org/)                      | [cloud_hypervisor/vm](https://github.com/usbarmory/tamago/tree/master/board/cloud_hypervisor/vm#executing-and-debugging) | VirtIO networking¹ |
| `firecracker`      | [Firecracker microvm](https://firecracker-microvm.github.io/)             | [firecracker/microvm](https://github.com/usbarmory/tamago/tree/master/board/firecracker/microvm#executing-and-debugging) | VirtIO networking¹ |
| `microvm`          | [QEMU microvm](https://www.qemu.org/docs/master/system/i386/microvm.html) | [qemu/microvm](https://github.com/usbarmory/tamago/tree/master/board/qemu/microvm#executing-and-debugging)               | VirtIO networking¹ |
| `gcp`              | [Google Compute Engine](https://cloud.google.com/products/compute)        | [tools](https://github.com/usbarmory/tamago-example/tree/master/tools)                                                   | VirtIO networking¹ |

¹ network configuration example in  _Emulated hardware with QEMU_

These targets are meant for paravirtualized execution as described in the
following sections.

Cloud Hypervisor
----------------

```
make example TARGET=cloud_hypervisor
cloud-hypervisor --kernel example --cpus boot=4 --memory size=4096M --net "tap=tap0" --serial tty --console off --seccomp false
```

Firecracker
-----------

Example shown via [firectl](https://github.com/firecracker-microvm/firectl):

```
make example TARGET=firecracker
firectl --kernel example --root-drive /dev/null --tap-device tap0/06:00:AC:10:00:01 -c 4 -m 4096
```

QEMU
----

```
make qemu TARGET=microvm SMP=4
```

Google Compute Engine
---------------------

The `gcp` target can be executed on [Google Compute Engine](https://cloud.google.com/products/compute), see
[tools](https://github.com/usbarmory/tamago-example/tree/master/tools) for more information.

It can also be executed in _Emulated hardware with QEMU_, when using `make
qemu` IP address 10.132.0.1/24 (VCP europe-west1) should be used on the host,
rather than 10.0.0.2/24.

Building and executing on ARM targets
=====================================

| `TARGET`    | Board                                                                                                                                                                                                          | Executing and debugging                                                                                  | Interface                                                                                |
|-------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------|
| `usbarmory` | [USB armory Mk II](https://github.com/usbarmory/usbarmory/wiki/Mk-II-Introduction)                                                                                                                             | [usbarmory](https://github.com/usbarmory/tamago/tree/master/board/usbarmory#executing-and-debugging)     | [Ethernet over USB](https://github.com/usbarmory/usbarmory/wiki/Host-communication)      |
| `usbarmory` | [USB armory Mk II LAN](https://github.com/usbarmory/usbarmory/wiki/Mk-II-LAN)                                                                                                                                  | [usbarmory](https://github.com/usbarmory/tamago/tree/master/board/usbarmory#executing-and-debugging)     | [Ethernet over USB](https://github.com/usbarmory/usbarmory/wiki/Host-communication), LAN |
| `mx6ullevk` | [MCIMX6ULL-EVK](https://www.nxp.com/design/design-center/development-boards-and-designs/i-mx-evaluation-and-development-boards/evaluation-kit-for-the-i-mx-6ull-and-6ulz-applications-processor:MCIMX6ULL-EVK) | [mx6ullevk](https://github.com/usbarmory/tamago/tree/master/board/nxp/mx6ullevk#executing-and-debugging) | LAN (2nd port)                                                                           |

Build the application executables as follows:

```
make imx TARGET=usbarmory
```

The targets support native (see relevant documentation links in the table above)
as well as emulated execution (e.g. `make qemu`).

Building and executing on ARM64 targets
=======================================

| `TARGET`    | Board                                                                                                    | Executing and debugging                                                                                  | Interface      |
|-------------|----------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------|----------------|
| `imx8mpevk` | [8MPLUSLPD4-EVK](https://www.nxp.com/design/design-center/development-boards-and-designs/8MPLUSLPD4-EVK) | [imx8mpevk](https://github.com/usbarmory/tamago/tree/master/board/nxp/imx8mpevk#executing-and-debugging) | LAN (1st port) |

> [!WARNING]
> This package is in early development stages, only emulated runs (qemu) have been tested.

Build and run the application executables as follows:

```
make qemu TARGET=imx8mpevk
```

Building and executing on RISCV64 targets
=========================================

| `TARGET`    | Board            | Executing and debugging                                                                                 | Interface       |
|-------------|------------------|---------------------------------------------------------------------------------------------------------|-----------------|
| `sifive_u`  | QEMU sifive_u    | [sifive_u](https://github.com/usbarmory/tamago/tree/master/board/qemu/sifive_u#executing-and-debugging) | Serial console  |

Build and run the application executables as follows:

```
make qemu TARGET=sifive_u
```

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
ip tuntap add dev tap0 mode tap group <your user group>
ip addr add 10.0.0.2/24 dev tap0
ip link set tap0 up
```

> [!NOTE]
> `TARGET=gcp make qemu` requires 10.132.0.1/24 (Google VCP europe-west1) as IP
> address, instead of 10.0.0.2/24

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
Copyright (c) The TamaGo Authors. All Rights Reserved.

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/tamago-example/blob/master/LICENSE) file.

The TamaGo logo is adapted from the Go gopher designed by Renee French and
licensed under the Creative Commons 3.0 Attributions license. Go Gopher vector
illustration by Hugo Arganda.
