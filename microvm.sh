make clean && TARGET=microvm make example

DESC=$(readelf -h example | grep "Entry point address"|grep -o -E "[a-f0-9]{8}$"|sed -e 's/\(..\)\(..\)\(..\)\(..\)/\\x\4\\x\3\\x\2\\x\1/')

echo -e -n "\x04\x00\x00\x00" >  note.bin # namesz
echo -e -n "\x04\x00\x00\x00" >> note.bin # descsz
echo -e -n "\x12\x00\x00\x00" >> note.bin # type
echo -e -n "Go \x00"          >> note.bin # name
echo -e -n "$DESC"            >> note.bin # desc (** change me to Entry point address **)

#objcopy --add-section .note.phv_start_addr=note.bin example
objcopy --update-section .note.go.buildid=note.bin example

echo "launching qemu"

OPTS=""

if [ "$1" == "gdb" ]; then
  OPTS="-S -s"
fi

qemu-system-x86_64 \
  -machine microvm,x-option-roms=on,pit=off,pic=off,rtc=on \
  -global virtio-mmio.force-legacy=false \
  -enable-kvm -cpu host,invtsc=on,kvmclock=on -no-reboot \
  -m 1.25G -nographic -monitor none -serial stdio \
  -device virtio-net-device,netdev=net0 -netdev tap,id=net0,ifname=tap0,script=no,downscript=no \
  -kernel example $OPTS
