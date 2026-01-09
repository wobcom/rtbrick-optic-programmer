package rtbrick

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"log"
	"math"
	"strconv"
	"strings"
)

func ParseI2CDump(dump string) ([]byte, error) {

	lines := strings.Split(dump, "\n")

	buf := make([]byte, 0, 1024)
	w := bytes.NewBuffer(buf)

	for _, line := range lines[1:17] {

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
	I2CPage12Grid50Ghz = iota
	I2CPage12Grid100Ghz
)

var I2CPage12GridMap = map[I2CPage12Grid]byte{
	I2CPage12Grid50Ghz:  0x40,
	I2CPage12Grid100Ghz: 0x50,
}

var I2CPage12GridMultiplierMap = map[I2CPage12Grid]int{
	I2CPage12Grid50Ghz:  5,
	I2CPage12Grid100Ghz: 10,
}

var I2CPage12GridNameMap = map[I2CPage12Grid]int{
	I2CPage12Grid50Ghz:  50,
	I2CPage12Grid100Ghz: 100,
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
	gridSetting := keysByValue(I2CPage12GridMap, bitfieldGrid)

	u := binary.BigEndian.Uint16(dump[136:138])
	frequencyOffset := int(TwoComplement16(16, u))
	gridMultiplier := I2CPage12GridMultiplierMap[*gridSetting]

	opticFrequency := 19310 + (frequencyOffset * gridMultiplier)
	channelSetting := keysByValue(DWDMGridMap, opticFrequency)

	status := dump[231]

	return I2CPage12{
		Grid:            *gridSetting,
		GridDisplay:     strconv.Itoa(I2CPage12GridNameMap[*gridSetting]),
		FrequencyOffset: frequencyOffset,
		Frequency:       opticFrequency,
		Channel:         channelSetting,
		Status:          status,
	}
}

func GetGridProgramming(gridStr int) (page, byte int, value byte) {
	grid := keysByValue(I2CPage12GridNameMap, gridStr)
	newValue := I2CPage12GridMap[*grid]
	return 0x12, 128, newValue
}

func GetChannelProgramming(gridStr int, newChannel int) (page, byte int, value byte, page2, byte2 int, value2 byte) {
	gridSetting := keysByValue(I2CPage12GridNameMap, gridStr)

	targetFrequency := DWDMGridMap[newChannel]
	gridMultiplier := I2CPage12GridMultiplierMap[*gridSetting]

	targetOffset := (targetFrequency - 19310) / gridMultiplier

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
