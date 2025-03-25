DESC=$(readelf -h $1 | grep "Entry point address"|grep -o -E "[a-f0-9]{8}$"|sed -e 's/\(..\)\(..\)\(..\)\(..\)/\\x\4\\x\3\\x\2\\x\1/')

echo -e -n "\x04\x00\x00\x00" >  note.bin # namesz
echo -e -n "\x04\x00\x00\x00" >> note.bin # descsz
echo -e -n "\x12\x00\x00\x00" >> note.bin # type
echo -e -n "Go \x00"          >> note.bin # name
echo -e -n "$DESC"            >> note.bin # desc (** change me to Entry point address **)

objcopy --update-section .note.go.buildid=note.bin $1
