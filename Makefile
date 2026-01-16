# Copyright (c) The TamaGo Authors. All Rights Reserved.
#
# Use of this source code is governed by the license
# that can be found in the LICENSE file.

SHELL = /bin/bash

APP := example
TARGET ?= usbarmory
TEXT_START := 0x80010000 # ramStart (defined in mem.go under relevant tamago/soc package) + 0x10000
TAGS := $(TARGET)

ifeq ($(TARGET),$(filter $(TARGET), microvm gcp))

SMP ?= $(shell nproc)
TEXT_START := 0x10010000 # ramStart (defined in mem.go under tamago/amd64 package) + 0x10000
GOENV := GOOS=tamago GOARCH=amd64

ifeq ($(TARGET),microvm)

QEMU ?= qemu-system-x86_64 -machine microvm,x-option-roms=on,pit=off,pic=off,rtc=on \
        -smp $(SMP) \
        -global virtio-mmio.force-legacy=false \
        -enable-kvm -cpu host,invtsc=on,kvmclock=on -no-reboot \
        -m 4G -nographic -monitor none -serial stdio \
        -device virtio-net-device,netdev=net0 -netdev tap,id=net0,ifname=tap0,script=no,downscript=no

endif

ifeq ($(TARGET),gcp)

QEMU ?= qemu-system-x86_64 -machine q35,pit=off,pic=off \
        -smp $(SMP) \
        -enable-kvm -cpu host,invtsc=on,kvmclock=on -no-reboot \
        -m 4G -nographic -monitor none -serial stdio \
        -device pcie-root-port,port=0x10,chassis=1,id=pci.0,bus=pcie.0,multifunction=on,addr=0x3 \
        -device virtio-net-pci,netdev=net0,mac=42:01:0a:84:00:02,disable-modern=true -netdev tap,id=net0,ifname=tap0,script=no,downscript=no

QEMU-img ?= qemu-system-x86_64 -machine q35 -m 4G -smp $(SMP) \
            -machine accel=kvm:tcg -cpu max \
            -vga none -display none -serial stdio \
            -device pcie-root-port,port=0x10,chassis=1,id=pci.1,bus=pcie.0,multifunction=on,addr=0x3 \
            -device pcie-root-port,port=0x11,chassis=2,id=pci.2,bus=pcie.0,addr=0x3.0x1 \
            -device pcie-root-port,port=0x12,chassis=3,id=pci.3,bus=pcie.0,addr=0x3.0x2 \
            -device virtio-scsi-pci,bus=pci.2,addr=0x0,id=scsi0 \
            -device scsi-hd,bus=scsi0.0,drive=hd0 \
            -device isa-debug-exit \
            -device virtio-rng-pci \
            -device virtio-balloon \
            -device virtio-net-pci=netdev=net0,mac=42:01:0a:84:00:02,disable-modern=true -netdev tap,id=net0,ifname=tap0,script=no,downscript=no \
            -drive file=$(APP).img,format=raw,if=none,id=hd0

endif

endif

ifeq ($(TARGET),$(filter $(TARGET), firecracker cloud_hypervisor))
TEXT_START := 0x10010000 # ramStart (defined in mem.go under tamago/amd64 package) + 0x10000
GOENV := GOOS=tamago GOARCH=amd64
endif

ifeq ($(TARGET),sifive_u)
GOENV := GOOS=tamago GOARCH=riscv64
QEMU ?= qemu-system-riscv64 -machine sifive_u -m 512M \
        -nographic -monitor none -semihosting -serial stdio -net none \
        -dtb $(CURDIR)/qemu.dtb -bios $(CURDIR)/tools/bios.bin
endif

ifeq ($(TARGET),$(filter $(TARGET), imx8mpevk mx6ullevk))
UART1 := stdio
UART2 := null
NET   := nic,model=imx.enet,netdev=net0 -netdev tap,id=net0,ifname=tap0,script=no,downscript=no
TAGS  := $(TARGET),linkramsize
endif

ifeq ($(TARGET),usbarmory)
UART1 := null
UART2 := stdio
NET   := none
TAGS  := $(TARGET),linkramsize
endif

ifeq ($(TARGET),imx8mpevk)
TEXT_START := 0x40010000 # ramStart (defined in mem.go under tamago/soc package) + 0x10000
GOENV := GOOS=tamago GOARCH=arm64
QEMU ?= qemu-system-aarch64 -machine imx8mp-evk -m 512M -smp 1 \
        -nographic -monitor none -semihosting \
        -serial $(UART1) -serial $(UART2) -net $(NET)
endif

ifeq ($(TARGET), $(filter $(TARGET), mx6ullevk usbarmory))
GOENV := GOOS=tamago GOARM=7 GOARCH=arm
QEMU ?= qemu-system-arm -machine mcimx6ul-evk -cpu cortex-a7 -m 512M \
        -nographic -monitor none -semihosting \
        -serial $(UART1) -serial $(UART2) -net $(NET)
endif

GOFLAGS := -tags ${TAGS},native -trimpath -ldflags "-T $(TEXT_START) -R 0x1000"

.PHONY: clean qemu qemu-gdb

check_tamago:
	@if [ "${TAMAGO}" == "" ] || [ ! -f "${TAMAGO}" ]; then \
		echo 'You need to set the TAMAGO variable to a compiled version of https://github.com/usbarmory/tamago-go'; \
		exit 1; \
	fi

clean:
	@rm -fr $(APP) $(APP).bin $(APP).img $(APP).imx $(APP)-signed.imx $(APP).csf $(APP).dcd
	@rm -fr cmd/*.yaml qemu.dtb tools/bios.bin tools/mbr.bin tools/mbr.lst

#### generic targets ####

all: $(APP)

elf: $(APP)

qemu: GOFLAGS := $(GOFLAGS:native=semihosting)
qemu: $(APP)
	@if [ "${QEMU}" == "" ]; then \
		echo 'qemu not available for this target'; \
		exit 1; \
	fi
	$(QEMU) -kernel $(APP)

qemu-gdb: GOFLAGS := $(GOFLAGS:native=semihosting)
qemu-gdb: GOFLAGS := $(GOFLAGS:-w=)
qemu-gdb: GOFLAGS := $(GOFLAGS:-s=)
qemu-gdb: $(APP)
	$(QEMU) -kernel $(APP) -S -s

qemu-img: $(APP).img
	@if [ "${QEMU-img}" == "" ]; then \
		echo 'qemu-img not available for this target'; \
		exit 1; \
	fi
	$(QEMU-img)

qemu-img-gdb: $(APP).img
	$(QEMU-img) -S -s

#### AMD64 targets ####

ifeq ($(TARGET),$(filter $(TARGET), microvm firecracker cloud_hypervisor gcp))

$(APP): check_tamago
	$(GOENV) $(TAMAGO) build $(GOFLAGS) -o ${APP}

img: $(APP).img

$(APP).bin: $(APP)
	objcopy -j .text -j .rodata -j .shstrtab -j .typelink \
	    -j .itablink -j .gopclntab -j .go.buildinfo -j .noptrdata -j .data \
	    -j .bss --set-section-flags .bss=alloc,load,contents \
	    -j .noptrbss --set-section-flags .noptrbss=alloc,load,contents \
	    $(APP) -O binary $(APP).bin

tools/mbr.bin: tools/mbr.s $(APP) $(APP).bin
	cd $(CURDIR)/tools && ./build_mbr.sh ../$(APP) ../$(APP).bin $(TEXT_START)
	@if (( $$(stat -c %s tools/mbr.bin) != 512 )); then \
		echo "ERROR: tools/mbr.bin size != 512."; \
		exit 1; \
	fi

$(APP).img: $(APP).bin tools/mbr.bin
	dd if=/dev/zero of=$(APP).img bs=1M count=100 status=none
	dd if=tools/mbr.bin of=$(APP).img conv=notrunc status=none
	dd if=$(APP).bin of=$(APP).img bs=1024 seek=100 conv=notrunc status=none

endif

#### ARM targets ####

ifeq ($(TARGET),$(filter $(TARGET), mx6ullevk usbarmory))

$(APP): check_tamago IMX6UL.yaml IMX6ULL.yaml
	$(GOENV) $(TAMAGO) build $(GOFLAGS) -o ${APP}

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

endif

#### ARM64 targets ####

ifeq ($(TARGET),imx8mpevk)
$(APP): check_tamago IMX8MP.yaml
	$(GOENV) $(TAMAGO) build $(GOFLAGS) -o ${APP}
endif

IMX8MP.yaml: check_tamago
IMX8MP.yaml: GOMODCACHE=$(shell ${TAMAGO} env GOMODCACHE)
IMX8MP.yaml: CRUCIBLE_PKG=$(shell grep "github.com/usbarmory/crucible v" go.mod | awk '{print $$1"@"$$2}')
IMX8MP.yaml:
	${TAMAGO} install github.com/usbarmory/crucible/cmd/habtool@latest
	cp -f $(GOMODCACHE)/$(CRUCIBLE_PKG)/cmd/crucible/fusemaps/IMX8MP.yaml cmd/IMX8MP.yaml

#### RISCV64 targets ####

ifeq ($(TARGET),$(filter $(TARGET), sifive_u))

qemu.dtb: GOMODCACHE=$(shell ${TAMAGO} env GOMODCACHE)
qemu.dtb: TAMAGO_PKG=$(shell grep "github.com/usbarmory/tamago v" go.mod | awk '{print $$1"@"$$2}')
qemu.dtb:
	echo $(GOMODCACHE)
	echo $(TAMAGO_PKG)
	dtc -I dts -O dtb $(GOMODCACHE)/$(TAMAGO_PKG)/board/qemu/sifive_u/qemu-riscv64-sifive_u.dts -o $(CURDIR)/qemu.dtb 2> /dev/null

$(APP): check_tamago qemu.dtb
	$(GOENV) $(TAMAGO) build $(GOFLAGS) -o ${APP} && \
	RT0=$$(riscv64-linux-gnu-readelf -a $(APP)|grep -i 'Entry point' | cut -dx -f2) && \
	echo ".equ RT0_RISCV64_TAMAGO, 0x$$RT0" > $(CURDIR)/tools/bios.cfg && \
	cd $(CURDIR)/tools && ./build_riscv64_bios.sh

endif
