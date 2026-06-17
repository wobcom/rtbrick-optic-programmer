package util

import (
	"encoding/binary"
	"iter"
	"math"
)

func TwoComplement16(sizeBits uint8, data uint16) (v int16) {
	n := float64(sizeBits - 1)
	p := math.Pow(2, n) // 2^(N-1)
	mask := uint16(p)

	tmp1 := -int16(data & mask)
	tmp2 := int16(data & ^mask)

	v = tmp1 + tmp2
	return v
}

func ReadBeUint16(bin []byte, baseOffset byte) uint16 {
	// CMIS states that the default endianness for numerical data types is big endian
	// unless stated otherwise
	return binary.BigEndian.Uint16(bin[baseOffset : baseOffset+0x02]) // go slices use half-open range, like this: [a,b[
}

func ReadBeUint32(bin []byte, baseOffset byte) uint32 {
	return binary.BigEndian.Uint32(bin[baseOffset : baseOffset+0x04]) // go slices use half-open ranges
}

func ReadBeInt16(bin []byte, baseOffset byte) int16 {
	return TwoComplement16(16, ReadBeUint16(bin, baseOffset))
}

func ReadBeInt16AndShiftBase(bin []byte, baseOffset *byte) int16 {
	i := ReadBeInt16(bin, *baseOffset)
	*baseOffset += 0x02
	return i
}

func ReadBeUint16AndShiftBase(bin []byte, baseOffset *byte) uint16 {
	i := ReadBeUint16(bin, *baseOffset)
	*baseOffset += 0x02
	return i
}

// BinDiffIterator will yield offset and value for each new byte that is different from original byte. old and new must
// be of the exact same size. I didn't fit old and new value in there cos go only has support for up to Seq2 iter.
func BinDiffIterator(old []byte, new []byte) iter.Seq2[byte, byte] {
	return func(yield func(byte, byte) bool) {
		var i byte = 0x00
		for _, i = range old {
			if new[i] != old[i] {
				if !yield(i, new[i]) {
					return
				}
			}
		}
	}
}
