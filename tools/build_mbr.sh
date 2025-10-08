ELF=$1
BIN=$2
START=$3

OFFSET=200
SIZE=$(stat -c %s $BIN)
ENTRY=0x$(dd if=$ELF bs=1 count=4 skip=24 | xxd -e -g4 | xxd -r | xxd -p)

echo "TAMAGO_OFFSET = $OFFSET"
echo "TAMAGO_SIZE   = $SIZE"
echo "TAMAGO_ENTRY  = $ENTRY"
echo "TAMAGO_START  = $START"

nasm -l mbr.lst mbr.s -o mbr.bin \
    -dTAMAGO_OFFSET=$OFFSET \
    -dTAMAGO_SIZE=$SIZE \
    -dTAMAGO_START=$START \
    -dTAMAGO_ENTRY=$ENTRY
