package util

import (
	"encoding/binary"
	"testing"
)

const bufTemplate = "0x%2x%2x"

func TestTwoComplement16EncodeZero(t *testing.T) {
	buf := TwoComplement16Encode(0)
	if buf[0] != 0x00 || buf[1] != 0x00 {
		t.Errorf(bufTemplate, buf[0], buf[1])
	}
}

func TestTwoComplement16EncodePositive(t *testing.T) {
	buf := TwoComplement16Encode(32767)
	if buf[0] != 0x7F || buf[1] != 0xFF {
		t.Errorf(bufTemplate, buf[0], buf[1])
	}
}

func TestTwoComplement16EncodeNegative(t *testing.T) {
	buf := TwoComplement16Encode(-32768)
	if buf[0] != 0x80 || buf[1] != 0x00 {
		t.Errorf(bufTemplate, buf[0], buf[1])
	}
}
func TestTwoComplement16EncodeNegative2(t *testing.T) {
	buf := TwoComplement16Encode(-11)
	if buf[0] != 0xFF || buf[1] != 0xF5 {
		t.Errorf(bufTemplate, buf[0], buf[1])
	}
}

func TestTwoComplement16ReadZero(t *testing.T) {
	buf := TwoComplement16Encode(0)

	data := binary.BigEndian.Uint16(buf)
	n := TwoComplement16(16, data)

	if n != 0 {
		t.Errorf(bufTemplate, buf[0], buf[1])
	}
}

func TestTwoComplement16ReadPositive(t *testing.T) {
	buf := TwoComplement16Encode(32767)

	data := binary.BigEndian.Uint16(buf)
	n := TwoComplement16(16, data)

	if n != 32767 {
		t.Errorf(bufTemplate, buf[0], buf[1])
	}
}

func TestTwoComplement16ReadNegative(t *testing.T) {
	buf := TwoComplement16Encode(-32768)

	data := binary.BigEndian.Uint16(buf)
	n := TwoComplement16(16, data)

	if n != -32768 {
		t.Errorf(bufTemplate, buf[0], buf[1])
	}
}

func TestBinDiffIterator(t *testing.T) {
	buf := []byte{
		/*     00    01    02    03    04    05    06    07    08    09    0a    0b    0c    0d    0e    0f*/
		/*00*/ 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		/*10*/ 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		/*20*/ 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		/*30*/ 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		/*40*/ 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		/*50*/ 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		/*60*/ 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	bufCopy := CopySlice(buf)

	bufCopy[0x00] = 0xFF
	bufCopy[0x6f] = 0xFF

	var offsets []int
	var values []byte

	for offset, value := range BinDiffIterator(buf, bufCopy) {
		offsets = append(offsets, offset)
		values = append(values, value)
	}

	if !(offsets[0] == 0x00 && values[0] == 0xFF) {
		t.Errorf(bufTemplate, buf[0], buf[1])
	}
	if !(offsets[1] == 0x6F && values[1] == 0xFF) {
		t.Errorf(bufTemplate, buf[0], buf[1])
	}
}
