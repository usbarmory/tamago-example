make clean && TARGET=microvm make example

echo -e -n "\x04\x00\x00\x00" >  note.bin # namesz
echo -e -n "\x04\x00\x00\x00" >> note.bin # descsz
echo -e -n "\x12\x00\x00\x00" >> note.bin # type
echo -e -n "Go \x00"          >> note.bin # name
echo -e -n "\x80\xd4\x07\x10" >> note.bin # desc (** change me to Entry point address **)

#objcopy --add-section .note.phv_start_addr=note.bin example
objcopy --update-section .note.go.buildid=note.bin example

echo "launching qemu"

OPTS=""

if [ "$1" == "gdb" ]; then
  OPTS="-S -s"
fi

qemu-system-x86_64 \
  -machine microvm -enable-kvm -cpu host \
  -m 1G -nographic -monitor none -serial stdio -kernel example $OPTS \
  -device virtio-net-device,netdev=net0 -netdev tap,id=net0,ifname=tap0,script=no,downscript=no
  # -net none
  #-m 1G -nographic -monitor none -serial stdio -net none -kernel example -d exec,nochain,cpu,in_asm
