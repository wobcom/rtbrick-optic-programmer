package rtbrick

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"log/slog"
	"math"
	"strconv"
	"strings"
)

func ParseI2CDump(dump string) ([]byte, error) {
	slog.Debug("======== I2C Dump ========")
	slog.Debug("\n" + dump)
	slog.Debug("======== ======== ========")

	lines := strings.Split(dump, "\n")

	buf := make([]byte, 0, 1024)
	w := bytes.NewBuffer(buf)

	for _, line := range lines[1:17] {

		p1 := strings.Split(line, ": ")[1]
		p2 := strings.Split(p1, "    ")[0]

		for _, x := range strings.Split(p2, " ") {
			b, err := hex.DecodeString(x)
			if err != nil {
				slog.Error("could not parse", "code", err)
				panic(err)
			}
			w.Write(b)
		}

	}

	allBytes := w.Bytes()
	slog.Debug("raw_decoded_i2c_bytes", slog.String("hex_string", hex.EncodeToString(allBytes)))
	return allBytes, nil

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

func ToTwoComplement16(a int16) []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(a))
	return b
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
	I2CPage12Grid3p125 = iota
	I2CPage12Grid6p25
	I2CPage12Grid12p5
	I2CPage12Grid25
	I2CPage12Grid50
	I2CPage12Grid100
	I2CPage12Grid33
	I2CPage12Grid75
	I2CPage12GridRESERVED
	I2CPage12GridUNAVAILABLE
)

type DefaultMap[K comparable, V comparable] struct {
	m map[K]V
	d V
}

func (m DefaultMap[K, V]) get(key K) V {
	if v, ok := m.m[key]; ok {
		return v
	}
	return m.d
}

func (m DefaultMap[K, V]) getKey(val V) K {
	for k, v := range m.m {
		if v == val {
			return k
		}
	}
	panic("value not found")
}

// I2CPage12GridMap cmis 12h reg 128-135  grid spacing
var I2CPage12GridMap = DefaultMap[byte, I2CPage12Grid]{
	m: map[byte]I2CPage12Grid{
		0b0000_0000: I2CPage12Grid3p125,
		0b0010_0000: I2CPage12Grid12p5,
		0b0001_0000: I2CPage12Grid6p25,
		0b0011_0000: I2CPage12Grid25,
		0b0100_0000: I2CPage12Grid50,
		0b0101_0000: I2CPage12Grid100,
		0b0110_0000: I2CPage12Grid33,
		0b0111_0000: I2CPage12Grid75,
		0b1111_0000: I2CPage12GridUNAVAILABLE,
	},
	d: I2CPage12GridRESERVED,
}

// I2CPage12GridMultiplierMap  Hz
var I2CPage12GridMultiplierMap = map[I2CPage12Grid]int{
	I2CPage12Grid3p125: 3.125e9,
	I2CPage12Grid6p25:  6.25e9,
	I2CPage12Grid12p5:  12.5e9,
	I2CPage12Grid25:    25.0e9,
	I2CPage12Grid50:    50.0e9,
	I2CPage12Grid100:   100.0e9,
	I2CPage12Grid33:    33.0e9,
	I2CPage12Grid75:    75.0e9,
}

// I2CPage12GridNameMap  Ghz
var I2CPage12GridNameMap = map[I2CPage12Grid]float64{
	I2CPage12Grid3p125: 3.125,
	I2CPage12Grid6p25:  6.25,
	I2CPage12Grid12p5:  12.5,
	I2CPage12Grid25:    25.0,
	I2CPage12Grid50:    50.0,
	I2CPage12Grid100:   100.0,
	I2CPage12Grid33:    33.0,
	I2CPage12Grid75:    75.0,
}

type I2CPage12 struct {
	Grid            I2CPage12Grid
	GridDisplay     string
	FrequencyOffset int
	Frequency       int
	Channel         *int
	Status          byte
}

func InterpretPage12(dump []byte) I2CPage12 {
	bitfieldGrid := dump[128]
	gridSetting := I2CPage12GridMap.get(bitfieldGrid)

	u := binary.BigEndian.Uint16(dump[136:138])
	frequencyOffset := int(TwoComplement16(16, u))
	gridMultiplier := I2CPage12GridMultiplierMap[gridSetting]

	opticFrequency := DWDMCenterFreqHz + (frequencyOffset * gridMultiplier)
	channelSetting := keysByValue(DWDMGridMap, opticFrequency)

	status := dump[231]

	return I2CPage12{
		Grid:            gridSetting,
		GridDisplay:     strconv.FormatFloat(I2CPage12GridNameMap[gridSetting], 'f', 3, 64),
		FrequencyOffset: frequencyOffset,
		Frequency:       opticFrequency,
		Channel:         channelSetting,
		Status:          status,
	}
}

func GetGridProgramming(gridStr float64) (page, byte int, value byte) {
	grid := keysByValue(I2CPage12GridNameMap, gridStr)
	newValue := I2CPage12GridMap.getKey(*grid)
	return 0x12, 128, newValue
}

func GetChannelProgramming(gridStr float64, newChannel int) (page, byte int, value byte, page2, byte2 int, value2 byte) {
	gridSetting := keysByValue(I2CPage12GridNameMap, gridStr)

	targetFrequency := DWDMGridMap[newChannel]
	gridMultiplier := I2CPage12GridMultiplierMap[*gridSetting]

	targetOffset := (targetFrequency - DWDMCenterFreqHz) / gridMultiplier

	sendBytes := ToTwoComplement16(int16(targetOffset))

	return 0x12, 137, sendBytes[0], 0x12, 136, sendBytes[1]
}

type I2CPage1E struct {
	FlexTuneEnabled    bool
	PowerClassOverride uint8
}

func InterpretPage1E(dump []byte) I2CPage1E {

	flexTuneEnabled := false
	if dump[200] == 0x01 {
		flexTuneEnabled = true
	}

	return I2CPage1E{
		FlexTuneEnabled:    flexTuneEnabled,
		PowerClassOverride: dump[253],
	}
}

func GetFlexTuneProgramming() (page, byte int, value byte) {
	var flexTuneBit uint8 = 0b00000000

	return 0x1E, 200, flexTuneBit
}

func GetPowerClassProgramming() (page, byte int, value byte) {
	var powerClassBit uint8 = 0x01

	return 0x1E, 253, powerClassBit
}

type I2CPageB0 struct {
	NominalWavelengthControlEnabled bool
}

func InterpretPageB0(dump []byte) I2CPageB0 {

	nominalWavelengthControlEnabled := false
	if dump[129] == 0x01 {
		nominalWavelengthControlEnabled = true
	}

	return I2CPageB0{
		NominalWavelengthControlEnabled: nominalWavelengthControlEnabled,
	}
}

func GetNominalWavelengthControlProgramming() (page, byte int, value byte) {
	var enableBit uint8 = 0b00000001
	return 0xB0, 129, enableBit
}

type I2CPage00 struct {
	VendorName   string
	VendorPN     string
	VendorSN     string
	LowPowerMode bool
}

func InterpretPage00(dump []byte) I2CPage00 {

	bit99Bitmask := Bitmask(dump[99])

	isLowPowerMode := bit99Bitmask.Has(Bit1)

	return I2CPage00{
		LowPowerMode: isLowPowerMode,
		VendorName:   ParseASCIIToString(dump[148:164]),
		VendorPN:     ParseASCIIToString(dump[168:184]),
		VendorSN:     ParseASCIIToString(dump[196:212]),
	}
}

func GetLowPowerProgramming(enableLowPower bool) (page, byte int, value byte) {
	var powerClassBit uint8 = 0b00000100
	if enableLowPower {
		powerClassBit = 0b00000010
	}

	return 0, 93, powerClassBit
}

func GetSoftReboot() (page, byte int, value byte) {
	var softRebootBit uint8 = 0b10000000

	return 0, 93, softRebootBit
}
