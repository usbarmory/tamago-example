# http://github.com/usbarmory/tamago-example
#
# Copyright (c) WithSecure Corporation
# https://foundry.withsecure.com
#
# Use of this source code is governed by the license
# that can be found in the LICENSE file.

BUILD_USER ?= $(shell whoami)
BUILD_HOST ?= $(shell hostname)
BUILD_DATE ?= $(shell /bin/date -u "+%Y-%m-%d %H:%M:%S")
BUILD = ${BUILD_USER}@${BUILD_HOST} on ${BUILD_DATE}
REV = $(shell git rev-parse --short HEAD 2> /dev/null)

SHELL = /bin/bash

APP := example
TARGET ?= usbarmory
TEXT_START := 0x80010000 # ramStart (defined in mem.go under relevant tamago/soc package) + 0x10000
TAGS := $(TARGET)

ifeq ($(TARGET),microvm)
TEXT_START := 0x10010000 # ramStart (defined in mem.go under relevant tamago/soc package) + 0x10000
GOENV := GOOS=tamago GOARCH=amd64
QEMU ?= qemu-system-x86_64 -machine microvm,x-option-roms=on,pit=off,pic=off,rtc=on \
        -global virtio-mmio.force-legacy=false \
        -enable-kvm -cpu host,invtsc=on,kvmclock=on -no-reboot \
        -m 4G -nographic -monitor none -serial stdio \
        -device virtio-net-device,netdev=net0 -netdev tap,id=net0,ifname=tap0,script=no,downscript=no
endif

ifeq ($(TARGET),sifive_u)
GOENV := GOOS=tamago GOARCH=riscv64
QEMU ?= qemu-system-riscv64 -machine sifive_u -m 512M \
        -nographic -monitor none -semihosting -serial stdio -net none \
        -dtb $(CURDIR)/qemu.dtb -bios $(CURDIR)/tools/bios.bin
endif

ifeq ($(TARGET),$(filter $(TARGET), mx6ullevk usbarmory))

TAGS := $(TARGET),linkramsize

ifeq ($(TARGET),mx6ullevk)
UART1 := stdio
UART2 := null
NET   := nic,model=imx.enet,netdev=net0 -netdev tap,id=net0,ifname=tap0,script=no,downscript=no
endif

ifeq ($(TARGET),usbarmory)
UART1 := null
UART2 := stdio
NET   := none
endif

GOENV := GOOS=tamago GOARM=7 GOARCH=arm
QEMU ?= qemu-system-arm -machine mcimx6ul-evk -cpu cortex-a7 -m 512M \
        -nographic -monitor none -semihosting \
        -serial $(UART1) -serial $(UART2) -net $(NET)

endif

GOFLAGS := -tags ${TAGS},native -trimpath -ldflags "-T $(TEXT_START) -R 0x1000 -X 'main.Build=${BUILD}' -X 'main.Revision=${REV}'"

.PHONY: clean qemu qemu-gdb

check_tamago:
	@if [ "${TAMAGO}" == "" ] || [ ! -f "${TAMAGO}" ]; then \
		echo 'You need to set the TAMAGO variable to a compiled version of https://github.com/usbarmory/tamago-go'; \
		exit 1; \
	fi

clean:
	@rm -fr $(APP) $(APP).bin $(APP).imx $(APP)-signed.imx $(APP).csf $(APP).dcd cmd/IMX6UL*.yaml qemu.dtb tools/bios.bin

#### generic targets ####

all: $(APP)

elf: $(APP)

qemu: GOFLAGS := $(GOFLAGS:native=semihosting)
qemu: $(APP)
	$(QEMU) -kernel $(APP)

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

IMX6UL.yaml: check_tamago
IMX6UL.yaml: GOMODCACHE=$(shell ${TAMAGO} env GOMODCACHE)
IMX6UL.yaml: CRUCIBLE_PKG=$(shell grep "github.com/usbarmory/crucible v" go.mod | awk '{print $$1"@"$$2}')
IMX6UL.yaml:
	${TAMAGO} install github.com/usbarmory/crucible/cmd/habtool@latest
	cp -f $(GOMODCACHE)/$(CRUCIBLE_PKG)/cmd/crucible/fusemaps/IMX6UL.yaml cmd/IMX6UL.yaml

IMX6ULL.yaml: check_tamago
IMX6ULL.yaml: GOMODCACHE=$(shell ${TAMAGO} env GOMODCACHE)
IMX6ULL.yaml: CRUCIBLE_PKG=$(shell grep "github.com/usbarmory/crucible v" go.mod | awk '{print $$1"@"$$2}')
IMX6ULL.yaml:
	${TAMAGO} install github.com/usbarmory/crucible/cmd/habtool@latest
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

ifeq ($(TARGET),microvm)
$(APP): check_tamago
	$(GOENV) $(TAMAGO) build $(GOFLAGS) -o ${APP}
	cd $(CURDIR) && ./tools/add_pvh_elf_note.sh ${APP}
endif

ifeq ($(TARGET),sifive_u)
$(APP): check_tamago qemu.dtb
	$(GOENV) $(TAMAGO) build $(GOFLAGS) -o ${APP} && \
	RT0=$$(riscv64-linux-gnu-readelf -a $(APP)|grep -i 'Entry point' | cut -dx -f2) && \
	echo ".equ RT0_RISCV64_TAMAGO, 0x$$RT0" > $(CURDIR)/tools/bios.cfg && \
	cd $(CURDIR)/tools && ./build_riscv64_bios.sh
endif

ifeq ($(TARGET),$(filter $(TARGET), mx6ullevk usbarmory))
$(APP): check_tamago IMX6UL.yaml IMX6ULL.yaml
	$(GOENV) $(TAMAGO) build $(GOFLAGS) -o ${APP}
endif

#### HAB secure boot ####

$(APP)-signed.imx: check_tamago check_hab_keys $(APP).imx
	${TAMAGO} install github.com/usbarmory/crucible/cmd/habtool@latest
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
