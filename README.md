TamaGo - bare metal Go for ARM SoCs - example application
=========================================================

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
applications on bare metal ARM System-on-Chip (SoC) components.

This example Go application illustrates use of the
[tamago](https://github.com/usbarmory/tamago) package
execute bare metal Go code on the following platforms:

| SoC          | Board                                                                                                                                                                                | SoC package                                                      | Board package                                                                         |
|--------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------|---------------------------------------------------------------------------------------|
| NXP i.MX6ULZ | [USB armory Mk II](https://github.com/usbarmory/usbarmory/wiki)                                                                                                                      | [imx6](https://github.com/usbarmory/tamago/tree/master/soc/imx6) | [usbarmory/mk2](https://github.com/usbarmory/tamago/tree/master/board/usbarmory)      |
| NXP i.MX6ULL | [MCIMX6ULL-EVK](https://www.nxp.com/design/development-boards/i-mx-evaluation-and-development-boards/evaluation-kit-for-the-i-mx-6ull-and-6ulz-applications-processor:MCIMX6ULL-EVK) | [imx6](https://github.com/usbarmory/tamago/tree/master/soc/imx6) | [nxp/mx6ullevk](https://github.com/usbarmory/tamago/tree/master/board/nxp/mx6ullevk)  |


Documentation
=============

For more information about TamaGo see its
[repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki).

Operation
=========

![Example screenshot](https://github.com/usbarmory/tamago/wiki/images/ssh.png)

The example application performs a variety of simple test procedures, each in
its separate goroutine:

  1. Directory and file write/read from an in-memory filesystem.

  2. SD/MMC card detection and read (only on non-emulated runs).

  3. Timer operation.

  4. Sleep operation.

  5. Random bytes collection (gathered from SoC TRNG on non-emulated runs).

  6. ECDSA signing and verification.

  7. Test BTC transaction creation and signing.

  8. Key derivation with SoC DCP (only on non emulated secure booted devices).

  9. Large memory allocation.

Once all tests are completed, and only on non-emulated hardware, the following
network services are started on [Ethernet over USB](https://github.com/usbarmory/usbarmory/wiki/Host-communication)
(ECM protocol, supported on Linux and macOS hosts).

  * SSH server on 10.0.0.1:22
  * HTTP server on 10.0.0.1:80
  * HTTPS server on 10.0.0.1:443

The web servers expose the following routes:

  * `/`: a welcome message
  * `/tamago-example.log`: log output
  * `/dir`: in-memory filesystem test directory
  * `/debug/pprof`: Go runtime profiling data through [pprof](https://golang.org/pkg/net/http/pprof/)
  * `/debug/statsviz`: Go runtime profiling data through [statsviz](https://github.com/arl/statsviz)

The SSH server exposes a basic shell with the following commands:

```
  help                                   # this help
  exit, quit                             # close session
  info                                   # SoC/board information
  rand                                   # gather 32 bytes from TRNG
  reboot                                 # reset the SoC/board
  stack                                  # stack trace of current goroutine
  stackall                               # stack trace of all goroutines
  date                                   # show   runtime date and time
  date <time in RFC3339 format>          # change runtime date and time
  dns  <fqdn>                            # resolve domain (requires routing)

  test                                   # launch example code

  ble                                    # enter BLE serial console
  i2c <n> <hex slave> <hex addr> <size>  # I²C bus read
  mmc <n> <hex offset> <size>            # internal MMC/SD card read
  md  <hex offset> <size>                # memory display (use with caution)
  mw  <hex offset> <hex value>           # memory write   (use with caution)
  led (white|blue) (on|off)              # LED control
  dcp <size> <sec>                       # benchmark hardware encryption
  otp <bank> <word>                      # OTP fuse display
```

Compiling
=========

Build the [TamaGo compiler](https://github.com/usbarmory/tamago-go)
(or use the [latest binary release](https://github.com/usbarmory/tamago-go/releases/latest)):

```
wget https://github.com/usbarmory/tamago-go/archive/refs/tags/latest.zip
unzip latest.zip
cd tamago-go-latest/src && ./all.bash
cd ../bin && export TAMAGO=`pwd`/go
```

Build the `example.imx` application executable:

```
git clone https://github.com/usbarmory/tamago-example
cd tamago-example && make CROSS_COMPILE=arm-none-eabi- TARGET=usbarmory imx
```

The supported targets for the `TARGET` environment variable are:
  * `usbarmory` - USB armory Mk II (default)
  * `mx6ullevk` - MCIMX6ULL-EVK

When cross compiling from a non-arm host, as shown in the example, ensure that
the `CROSS_COMPILE` variable is set according to the available toolchain (e.g.
`gcc-arm-none-eabi` package on Debian/Ubuntu).

The imx target also requires the `mkimage` tool from U-Boot (e.g.
`u-boot-tools` on Debian/Ubuntu).

Executing and debugging
=======================

Native hardware: imx image
--------------------------

Follow [these instructions](https://github.com/usbarmory/usbarmory/wiki/Boot-Modes-(Mk-II)#flashing-bootable-images-on-externalinternal-media)
using the built `example.imx` image.

Native hardware: existing bootloader
------------------------------------

Copy the built `example` binary on an external microSD card (replace `$dev`
with `0`) or the internal eMMC (replace `$dev` with `1`), then launch it from
the U-Boot console as follows:

```
ext2load mmc $dev:1 0x90000000 example
bootelf -p 0x90000000
```

For non-interactive execution modify the U-Boot configuration accordingly.

Standard output
---------------

The built in SSH server, once connected to, will redirect all logs to the
established session.

Alternatively the standard output can be accessed through the
[debug accessory](https://github.com/usbarmory/usbarmory/tree/master/hardware/mark-two-debug-accessory)
and the following `picocom` configuration:

```
picocom -b 115200 -eb /dev/ttyUSB2 --imap lfcrlf
```

Debugging
---------

The application can be debugged with GDB over JTAG using `openocd` and the
`imx6ull.cfg` and `gdbinit` debugging helpers published
[here](https://github.com/usbarmory/tamago/tree/master/_dev).

```
# start openocd daemon
openocd -f interface/ftdi/jtagkey.cfg -f imx6ull.cfg

# connect to the OpenOCD command line
telnet localhost 4444

# debug with GDB
arm-none-eabi-gdb -x gdbinit example
```

Hardware breakpoints can be set in the usual way:

```
hb ecdsa.Verify
continue
```

QEMU
----

The target can be executed under emulation as follows:

```
cd tamago-example && make qemu
```

The emulated target can be debugged with GDB by adding the `-S -s` flags to the
previous execution command, this will make qemu waiting for a GDB connection
that can be launched as follows:

```
arm-none-eabi-gdb -ex "target remote 127.0.0.1:1234" example
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
