# http://github.com/usbarmory/tamago-example
#
# Copyright (c) WithSecure Corporation
# https://foundry.withsecure.com
#
# Use of this source code is governed by the license
# that can be found in the LICENSE file.

BUILD_USER ?= $(shell whoami)
BUILD_HOST ?= $(shell hostname)
BUILD_DATE ?= $(shell /bin/date -u "+%Y-%m-%d,%H:%M:%S")
BUILD = ${BUILD_USER},${BUILD_HOST}:${BUILD_DATE}
REV = 22

SHELL = /bin/bash

APP := example
TARGET ?= "usbarmory"
TEXT_START := 0x80010000 # ramStart (defined in mem.go under relevant tamago/soc package) + 0x10000

ifeq ($(TARGET),sifive_u)

GOENV := GO_EXTLINK_ENABLED=0 CGO_ENABLED=0 GOOS=tamago GOARCH=riscv64
ENTRY_POINT := _rt0_riscv64_tamago
QEMU ?= qemu-system-riscv64 -machine sifive_u -m 1024M \
        -nographic -serial stdio -net none \
        -semihosting \
        -dtb $(CURDIR)/qemu.dtb \
        -bios $(CURDIR)/bios/bios.bin
else

ifeq ($(TARGET),mx6ullevk)
UART1 := stdio
UART2 := null
NET   := nic,model=imx.enet,netdev=net0 -netdev tap,id=net0,ifname=tap0,script=no,downscript=no
else
UART1 := null
UART2 := stdio
NET   := none
endif

GOENV := GO_EXTLINK_ENABLED=0 CGO_ENABLED=0 GOOS=tamago GOARM=7 GOARCH=arm
ENTRY_POINT := _rt0_arm_tamago
QEMU ?= qemu-system-arm -machine mcimx6ul-evk -cpu cortex-a7 -m 1024M \
        -nographic -monitor none -serial $(UART1) -serial $(UART2) -net $(NET) \
        -semihosting

endif

GOTAGS := -tags ${TARGET},linkramsize,native
GOLDFLAGSSTRIP := "-s -w -T $(TEXT_START) -E $(ENTRY_POINT) -R 0x1000"
GOLDFLAGSNOSTRIP := "-T $(TEXT_START) -E $(ENTRY_POINT) -R 0x1000"
GOFLAGS := $(GOTAGS) -trimpath -ldflags $(GOLDFLAGSNOSTRIP)

.PHONY: clean qemu qemu-gdb

check_tamago:
	@if [ "${TAMAGO}" == "" ] || [ ! -f "${TAMAGO}" ]; then \
		echo 'You need to set the TAMAGO variable to a compiled version of https://github.com/usbarmory/tamago-go'; \
		exit 1; \
	fi

check_uroot: check_tamago
	which u-root

clean:
	@rm -fr $(APP) $(APP).bin $(APP).imx $(APP)-signed.imx $(APP).csf $(APP).dcd cmd/IMX6ULL.yaml qemu.dtb bios/bios.bin

#### generic targets ####

all: $(APP)

elf: $(APP)

qemu: GOFLAGS := $(GOFLAGS:native=semihosting)
qemu: $(APP)
	$(QEMU) -kernel $(APP) -monitor /dev/ttys001

justqemu:
	$(QEMU) -kernel $(APP) -monitor /dev/ttys001

qemu-gdb: GOFLAGS := $(GOFLAGS:native=semihosting)
qemu-gdb: GOFLAGS := $(GOFLAGS:-w=)
qemu-gdb: GOFLAGS := $(GOFLAGS:-s=)
qemu-gdb: $(APP)
	$(QEMU) -kernel $(APP) -S -s

#### ARM targets ####

imx: $(APP).imx

imx_signed: $(APP)-signed.imx

check_hab_keys:
	@if [ "${HAB_KEYS}" == "" ]; then \
		echo 'You need to set the HAB_KEYS variable to the path of secure boot keys'; \
		echo 'See https://github.com/usbarmory/usbarmory/wiki/Secure-boot-(Mk-II)'; \
		exit 1; \
	fi

$(APP).bin: CROSS_COMPILE=arm-none-eabi-
$(APP).bin: $(APP)
	$(CROSS_COMPILE)objcopy -j .text -j .rodata -j .shstrtab -j .typelink \
	    -j .itablink -j .gopclntab -j .go.buildinfo -j .noptrdata -j .data \
	    -j .bss --set-section-flags .bss=alloc,load,contents \
	    -j .noptrbss --set-section-flags .noptrbss=alloc,load,contents \
	    $(APP) -O binary $(APP).bin

$(APP).imx: $(APP).bin $(APP).dcd
	mkimage -n $(APP).dcd -T imximage -e $(TEXT_START) -d $(APP).bin $(APP).imx
	# Copy entry point from ELF file
	dd if=$(APP) of=$(APP).imx bs=1 count=4 skip=24 seek=4 conv=notrunc

IMX6ULL.yaml: check_tamago
IMX6ULL.yaml: GOMODCACHE=$(shell ${TAMAGO} env GOMODCACHE)
IMX6ULL.yaml: CRUCIBLE_PKG=$(shell grep "github.com/usbarmory/crucible v" go.mod | awk '{print $$1"@"$$2}')
IMX6ULL.yaml:
	${TAMAGO} install github.com/usbarmory/crucible/cmd/habtool
	cp -f $(GOMODCACHE)/$(CRUCIBLE_PKG)/cmd/crucible/fusemaps/IMX6ULL.yaml cmd/IMX6ULL.yaml

$(APP).dcd: check_tamago
$(APP).dcd: GOMODCACHE=$(shell ${TAMAGO} env GOMODCACHE)
$(APP).dcd: TAMAGO_PKG=$(shell grep "github.com/usbarmory/tamago v" go.mod | awk '{print $$1"@"$$2}')
$(APP).dcd:
	@if test "${TARGET}" = "usbarmory"; then \
		cp -f $(GOMODCACHE)/$(TAMAGO_PKG)/board/usbarmory/mk2/imximage.cfg $(APP).dcd; \
	elif test "${TARGET}" = "mx6ullevk"; then \
		cp -f $(GOMODCACHE)/$(TAMAGO_PKG)/board/nxp/mx6ullevk/imximage.cfg $(APP).dcd; \
	else \
		echo "invalid target - options are: usbarmory, mx6ullevk"; \
		exit 1; \
	fi

#### RISC-V targets ####

qemu.dtb: GOMODCACHE=$(shell ${TAMAGO} env GOMODCACHE)
qemu.dtb: TAMAGO_PKG=$(shell grep "github.com/usbarmory/tamago v" go.mod | awk '{print $$1"@"$$2}')
qemu.dtb:
	echo $(GOMODCACHE)
	echo $(TAMAGO_PKG)
	dtc -I dts -O dtb $(GOMODCACHE)/$(TAMAGO_PKG)/board/qemu/sifive_u/qemu-riscv64-sifive_u.dts -o $(CURDIR)/qemu.dtb 2> /dev/null

#### application target ####

ifeq ($(TARGET),sifive_u)

$(APP): check_tamago qemu.dtb
	$(GOENV) $(TAMAGO) build $(GOFLAGS) -o ${APP} && \
	RT0=$$(riscv64-linux-gnu-readelf -a $(APP)|grep -i 'Entry point' | cut -dx -f2) && \
	echo ".equ RT0_RISCV64_TAMAGO, 0x$$RT0" > $(CURDIR)/bios/cfg.inc && \
	cd $(CURDIR)/bios && ./build.sh

else

$(APP): check_tamago IMX6ULL.yaml
	$(GOENV) $(TAMAGO) build $(GOFLAGS) -o ${APP}

uroot: check_uroot IMX6ULL.yaml $(APP)
	rm -rf $(PWD)/tdir
	$(GOENV) u-root -go-no-strip -no-strip -tmpdir $(PWD)/tdir -o tx -defaultsh="forth" -initcmd="forth" -gen-dir /tmp/x -uroot-source=~/go/src/github.com/u-root/u-root  $(GOTAGS)  -go-extra-args -ldflags="-T $(TEXT_START) -E $(ENTRY_POINT) -R 0x1000" .  \
	tamago \
	~/go/src/github.com/u-root/u-root/cmds/core/echo \
	~/go/src/github.com/u-root/u-root/cmds/exp/forth \
	~/go/src/github.com/u-root/u-root/cmds/core/wget
	mkdir -p bbin
	rm -f bbin/bb
	cpio -iv < tx  bbin/bb
	cp bbin/bb utx

builduroot:
	echo DO NOT DO THIS ANY MORE
	false
	cp uroot/init.go tdir/*/src/github.com/usbarmory/tamago-example/tamago
	echo the "" is required by BSD sed, which has its own wonderful rules.
	sed -i "" 's/os.Exit.0./if false { & } /' tdir/*/src/bb.u-root.com/bb/pkg/bbmain/register.go
	(cd tdir/*/src/bb.u-root.com/bb && $(GOENV) go build -o tx  $(GOTAGS)  \
			-ldflags="-T $(TEXT_START) -E $(ENTRY_POINT) -R 0x1000" .  )
	mkdir -p bbin
	rm -f bbin/bb
	cpio -iv < tx  bbin/bb
	cp tdir/*/src/bb.u-root.com/bb/tx utx

urootqemu: uroot
	$(QEMU) -kernel utx -monitor /dev/ttys001

endif

#### HAB secure boot ####

$(APP)-signed.imx: check_tamago check_hab_keys $(APP).imx
	${TAMAGO} install github.com/usbarmory/crucible/cmd/habtool
	$(shell ${TAMAGO} env GOPATH)/bin/habtool \
		-A ${HAB_KEYS}/CSF_1_key.pem \
		-a ${HAB_KEYS}/CSF_1_crt.pem \
		-B ${HAB_KEYS}/IMG_1_key.pem \
		-b ${HAB_KEYS}/IMG_1_crt.pem \
		-t ${HAB_KEYS}/SRK_1_2_3_4_table.bin \
		-x 1 \
		-s \
		-i $(APP).imx \
		-o $(APP).csf && \
	cat $(APP).imx $(APP).csf > $(APP)-signed.imx
