package rtbrick

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"log"
	"math"
	"strings"
)

func ParseI2CDump(dump string) ([]byte, error) {

	lines := strings.Split(dump, "\n")

	buf := make([]byte, 0, 1024)
	w := bytes.NewBuffer(buf)

	for _, line := range lines[1:16] {

		p1 := strings.Split(line, ": ")[1]
		p2 := strings.Split(p1, "    ")[0]

		for _, x := range strings.Split(p2, " ") {
			b, err := hex.DecodeString(x)
			if err != nil {
				log.Fatalf("could not parse, %v", err)
			}
			w.Write(b)
		}

	}

	return w.Bytes(), nil

}

func keysByValue[K comparable, V comparable](m map[K]V, value V) *K {
	for k, v := range m {
		if value == v {
			return &k
		}
	}
	return nil
}

func TwoComplement16(sizeBits uint8, data uint16) (v int16) {
	n := float64(sizeBits - 1)
	p := math.Pow(2, n) // 2^(N-1)
	mask := uint16(p)

	tmp1 := -int16(data & mask)
	tmp2 := int16(data & ^mask)

	v = tmp1 + tmp2
	return v
}

func ParseASCIIToString(part []byte) string {
	var asciiString string
	for _, code := range part {
		asciiString += string(rune(code))
	}
	return strings.TrimSpace(asciiString)
}

type I2CPage12Grid int

const (
	I2CPage12Grid50Ghz = iota
	I2CPage12Grid100Ghz
)

var I2CPage12GridMap = map[I2CPage12Grid]int{
	I2CPage12Grid50Ghz:  0x40,
	I2CPage12Grid100Ghz: 0x50,
}

var I2CPage12GridMultiplierMap = map[I2CPage12Grid]float32{
	I2CPage12Grid50Ghz:  0.05,
	I2CPage12Grid100Ghz: 0.1,
}

var I2CPage12GridNameMap = map[I2CPage12Grid]string{
	I2CPage12Grid50Ghz:  "50GHz",
	I2CPage12Grid100Ghz: "100GHz",
}

type I2CPage12 struct {
	Grid             I2CPage12Grid
	GridDisplay      string
	FrequencyOffset  int16
	FrequencyDisplay float32
}

func InterpretPage12(dump []byte) I2CPage12 {
	bitfieldGrid := int(dump[128])
	gridSetting := keysByValue(I2CPage12GridMap, bitfieldGrid)

	u := binary.BigEndian.Uint16(dump[136:138])
	frequencyOffset := TwoComplement16(16, u)

	gridMultiplier := I2CPage12GridMultiplierMap[*gridSetting]

	opticFrequency := 193.1 + (float32(frequencyOffset) * gridMultiplier)

	return I2CPage12{
		Grid:             *gridSetting,
		GridDisplay:      I2CPage12GridNameMap[*gridSetting],
		FrequencyOffset:  frequencyOffset,
		FrequencyDisplay: opticFrequency,
	}
}

type I2CPage1E struct {
	FlexTuneEnabled bool
	LowPowerMode    bool
}

func InterpretPage1E(dump []byte) I2CPage1E {

	flexTuneEnabled := false
	if dump[200] == 0x01 {
		flexTuneEnabled = true
	}

	bit99Bitmask := Bitmask(dump[99])

	isLowPowerMode := bit99Bitmask.Has(Bit1)

	return I2CPage1E{
		FlexTuneEnabled: flexTuneEnabled,
		LowPowerMode:    isLowPowerMode,
	}
}

type I2CPage1B struct {
	NominalWavelengthControlEnabled bool
}

func InterpretPage1B(dump []byte) I2CPage1B {

	nominalWavelengthControlEnabled := false
	if dump[129] == 0x01 {
		nominalWavelengthControlEnabled = true
	}

	return I2CPage1B{
		NominalWavelengthControlEnabled: nominalWavelengthControlEnabled,
	}
}

type I2CPage00 struct {
	VendorName string
	VendorPN   string
	VendorSN   string
}

func InterpretPage00(dump []byte) I2CPage00 {

	return I2CPage00{
		VendorName: ParseASCIIToString(dump[148:164]),
		VendorPN:   ParseASCIIToString(dump[168:184]),
		VendorSN:   ParseASCIIToString(dump[196:212]),
	}
}
