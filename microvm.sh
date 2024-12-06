make clean && TARGET=microvm make example

echo -e -n "\x04\x00\x00\x00" >  note.bin # namesz
echo -e -n "\x04\x00\x00\x00" >> note.bin # descsz
echo -e -n "\x12\x00\x00\x00" >> note.bin # type
echo -e -n "Go \x00"          >> note.bin # name
echo -e -n "\xa0\xf8\x07\x10" >> note.bin # desc

#objcopy --add-section .note.phv_start_addr=note.bin example
objcopy --update-section .note.go.buildid=note.bin example

echo "launching qemu"

qemu-system-x86_64 \
  -machine microvm \
  -m 1G -nographic -monitor none -serial stdio -net none -kernel example -S -s
  #-m 1G -nographic -monitor none -serial stdio -net none -kernel example -d exec,nochain,cpu,in_asm # -S -s
