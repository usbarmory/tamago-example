TamaGo - bare metal Go for ARM SoCs - USB armory example
========================================================

tamago | https://github.com/f-secure-foundry/tamago  

Copyright (c) F-Secure Corporation  
https://foundry.f-secure.com

![TamaGo gopher](https://github.com/f-secure-foundry/tamago/wiki/images/tamago.svg?sanitize=true)

Authors
=======

Andrea Barisani  
andrea.barisani@f-secure.com | andrea@inversepath.com  

Andrej Rosano  
andrej.rosano@f-secure.com   | andrej@inversepath.com  

Introduction
============

TamaGo is a project that aims to provide compilation and execution of
unencumbered Go applications for bare metal ARM System-on-Chip (SoC)
components.

This example Go application illustrates use of the
[usbarmory](https://github.com/f-secure-foundry/tamago/tree/master/usbarmory) package
part of [TamaGo](https://github.com/f-secure-foundry/tamago), to execute bare metal Go code on the
[USB armory](https://github.com/f-secure-foundry/usbarmory/wiki) Single Board Computer.

Documentation
=============

For more information about TamaGo see its
[repository](https://github.com/f-secure-foundry/tamago) and
[project wiki](https://github.com/f-secure-foundry/tamago/wiki).

Operation
=========

The example application performs a variety of simple test procedures, each in
its separate goroutine:

  1. Directory and file writt/read from an in-memory filesystem.

  2. Timer operation.

  3. Sleep operation.

  4. Random bytes collection (gathered from SoC TRNG on non-emulated runs).

  5. ECDSA signing and verification.

  6. Test BTC transaction creation and signing.

  7. Key derivation with DCP HSM (only on non-emulated runs).

  8. Large memory allocation.

Once all tests are completed, and only on non-emulated hardware, the following
network services are started on Ethernet over USB (ECM protocol, only supported
on Linux hosts).

  * UDP echo server on 10.0.0.1:1234
  * HTTP server on 10.0.0.1:80
  * HTTPS server on 10.0.0.1:443

The HTTP/HTTPS servers expose the following routes:

  * `/`: a welcome message
  * `/dir`: in-memory filesystem
  * `/debug/pprof`: Go runtime profiling data through [pprof](https://golang.org/pkg/net/http/pprof/)
  * `/debug/charts`: Go runtime profiling data through [debugcharts](https://github.com/mkevac/debugcharts)

Compiling
=========

Build the [TamaGo compiler](https://github.com/f-secure-foundry/tamago-go):

```
git clone https://github.com/f-secure-foundry/tamago-go -b tamago1.13.6
cd tamago-go/src && ./all.bash
cd ../bin && export TAMAGO=`pwd`/go
```

Build the example application:

```
git clone https://github.com/f-secure-foundry/tamago-example
cd tamago-example && make
```

Executing and debugging
=======================

Native hardware
---------------

Copy the compiled application on an external microSD card (replace `$dev` with
`0`) or the internal eMMC (replace `$dev` with `1`), then launch it from the
U-Boot console as follows:

```
ext2load mmc $dev:1 0x90000000 example
bootelf -p 0x90000000
```

For non-interactive execution modify the U-Boot configuration accordingly.

The standard output can be accessed through the
[debug accessory](https://github.com/f-secure-foundry/usbarmory/tree/master/hardware/mark-two-debug-accessory)
and the following `picocom` configuration:

```
picocom -b 115200 -eb /dev/ttyUSB2 --imap lfcrlf
```

The application can be debugged with GDB over JTAG using `openocd` and the
`imx6ull.cfg` and `gdbinit` debugging helpers published
[here](https://github.com/f-secure-foundry/tamago/tree/master/dev).

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

tamago | https://github.com/f-secure-foundry/tamago  
Copyright (c) F-Secure Corporation

This program is free software: you can redistribute it and/or modify it under
the terms of the GNU General Public License as published by the Free Software
Foundation under version 3 of the License.

This program is distributed in the hope that it will be useful, but WITHOUT ANY
WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
PARTICULAR PURPOSE. See the GNU General Public License for more details.

See accompanying LICENSE file for full details.

The TamaGo logo is adapted from the Go gopher designed by Renee French and
licensed under the Creative Commons 3.0 Attributions license. Go Gopher vector
illustration by Hugo Arganda.
