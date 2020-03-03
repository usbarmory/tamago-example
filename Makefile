# http://github.com/f-secure-foundry/tamago-example
#
# Copyright (c) F-Secure Corporation
# https://foundry.f-secure.com
#
# Use of this source code is governed by the license
# that can be found in the LICENSE file.

BUILD_USER = $(shell whoami)
BUILD_HOST = $(shell hostname)
BUILD_DATE = $(shell /bin/date -u "+%Y-%m-%d %H:%M:%S")
BUILD = ${BUILD_USER}@${BUILD_HOST} on ${BUILD_DATE}
REV = $(shell git rev-parse --short HEAD 2> /dev/null)

APP := example
GOENV := GO_EXTLINK_ENABLED=0 CGO_ENABLED=0 GOOS=tamago GOARM=7 GOARCH=arm
TEXT_START := 0x80010000 # ramStart (defined in imx6/imx6ul/memory.go) + 0x10000
GOFLAGS := -ldflags "-T $(TEXT_START) -E _rt0_arm_tamago -R 0x1000 -X 'main.Build=${BUILD}' -X 'main.Revision=${REV}'"
QEMU ?= qemu-system-arm -machine mcimx6ul-evk -cpu cortex-a7 -m 512M \
        -nographic -monitor none -serial null -serial stdio -net none \
        -semihosting -d unimp

SHELL = /bin/bash
UBOOT_VER=2019.07
USBARMORY_REPO=https://raw.githubusercontent.com/f-secure-foundry/usbarmory/master
LOSETUP_DEV=$(shell /sbin/losetup -f)
DISK_SIZE = 50MiB
JOBS=2
# microSD: 0, eMMC: 1
BOOTDEV ?= 0
BOOTCOMMAND = ext2load mmc $(BOOTDEV):1 0x90000000 example; bootelf -p 0x90000000

.PHONY: clean qemu qemu-gdb

all: $(APP)

$(APP):
	@if [ "${TAMAGO}" == "" ] || [ ! -f "${TAMAGO}" ]; then \
		echo 'You need to set the TAMAGO variable to a compiled version of https://github.com/f-secure-foundry/tamago-go'; \
		exit 1; \
	fi

	$(GOENV) $(TAMAGO) build $(GOFLAGS) -o ${APP}

$(APP).raw: $(APP) u-boot
	@if [ ! -f "$(APP).raw" ]; then \
		truncate -s $(DISK_SIZE) $(APP).raw && \
		sudo /sbin/parted $(APP).raw --script mklabel msdos && \
		sudo /sbin/parted $(APP).raw --script mkpart primary ext4 5M 100% && \
		sudo /sbin/losetup $(LOSETUP_DEV) $(APP).raw -o 5242880 --sizelimit $(DISK_SIZE) && \
		sudo /sbin/mkfs.ext4 -F $(LOSETUP_DEV) && \
		sudo /sbin/losetup -d $(LOSETUP_DEV) && \
		mkdir -p rootfs && \
		sudo mount -o loop,offset=5242880 -t ext4 $(APP).raw rootfs/ && \
		sudo cp ${APP} rootfs/ && \
		sudo umount rootfs && \
		sudo dd if=u-boot-${UBOOT_VER}/u-boot-dtb.imx of=$(APP).raw bs=512 seek=2 conv=fsync conv=notrunc && \
		rmdir rootfs; \
	fi

elf: $(APP)

raw: $(APP).raw

qemu: $(APP)
	$(QEMU) -kernel $(APP)

qemu-gdb: $(APP)
	$(QEMU) -kernel $(APP) -S -s

clean:
	rm -f $(APP)
	@rm -fr $(APP).raw u-boot-${UBOOT_VER}*

u-boot-${UBOOT_VER}.tar.bz2:
	wget ftp://ftp.denx.de/pub/u-boot/u-boot-${UBOOT_VER}.tar.bz2 -O u-boot-${UBOOT_VER}.tar.bz2
	wget ftp://ftp.denx.de/pub/u-boot/u-boot-${UBOOT_VER}.tar.bz2.sig -O u-boot-${UBOOT_VER}.tar.bz2.sig

u-boot-${UBOOT_VER}/u-boot.bin: u-boot-${UBOOT_VER}.tar.bz2
	gpg --verify u-boot-${UBOOT_VER}.tar.bz2.sig
	tar xf u-boot-${UBOOT_VER}.tar.bz2
	cd u-boot-${UBOOT_VER} && make distclean
	cd u-boot-${UBOOT_VER} && \
		wget ${USBARMORY_REPO}/software/u-boot/0001-ARM-mx6-add-support-for-USB-armory-Mk-II-board.patch && \
		wget ${USBARMORY_REPO}/software/u-boot/0001-Drop-linker-generated-array-creation-when-CONFIG_CMD.patch && \
		patch -p1 < 0001-ARM-mx6-add-support-for-USB-armory-Mk-II-board.patch && \
		patch -p1 < 0001-Drop-linker-generated-array-creation-when-CONFIG_CMD.patch && \
		make usbarmory-mark-two_defconfig; \
		sed -i -e 's/run start_normal/${BOOTCOMMAND}/' include/configs/usbarmory-mark-two.h
	cd u-boot-${UBOOT_VER} && CROSS_COMPILE=arm-linux-gnueabihf- ARCH=arm make -j${JOBS}

u-boot: u-boot-${UBOOT_VER}/u-boot.bin
