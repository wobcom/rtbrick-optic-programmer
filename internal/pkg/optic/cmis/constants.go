package cmis

const (
	BankSelectRegisterAddress = 0x7E
	PageSelectRegisterAddress = 0x7F
)

const (
	GridSupported75GhzMask = 0b1000_0000 >> iota
	GridSupported33GhzMask
	GridSupported100GhzMask
	GridSupported50GhzMask
	GridSupported25GhzMask
	GridSupported12p5GhzMask
	GridSupported6p25GhzMask
	GridSupported3p125GhzMask
)
const (
	FineTuningSupportedMask = 0b1000_0000 >> iota
	GridSupported150GhzMask
)

const (
	SoftwareResetMask               = 0b0000_1000
	LowPwrRequestMask               = 0b0001_0000
	ProgOutputPowerPerLaneSupported = 0b1000_0000
	MemoryModelMask                 = 0b1000_0000
	SteppedConfigOnlyMask           = 0b0100_0000
	I2CMciMaxSpeedMask              = 0b0011_0000
	SPIMCIMaxSpeedMask              = 0b0000_1100
	AutoCommissioningMask           = 0b0000_0011
	ModulePowerClassMask            = 0b1110_0000
	FarEndConfigurationMask         = 0b0001_0000
	BanksSupportedMask              = 0b0000_0011
	ExtraLaneBanksSupportedMask     = 0b0001_1111
)

const (
	WavelengthIsControllableMask      = 0b1000_0000
	TransmitterIsTunableMask          = 0b0100_0000
	SquelchMethodTxMask               = 0b0011_0000
	ForcedSquelchTxSupportedMask      = 0b0000_1000
	AutoSquelchDisableTxSupportedMask = 0b0000_0100
	OutputDisableTxSupportedMask      = 0b0000_0010
	InputPolarityFlipTxSupportedMask  = 0b0000_0001
	BankBroadcastSupportedMask        = 0b1000_0000
	AutoSquelchDisableRxSupportedMask = 0b0000_0100
	OutputDisableRxSupportedMask      = 0b0000_0010
	OutputPolarityFlipRxSupportedMask = 0b0000_0001
)

const TunableLaserControlStatusPage = 0x12
const MaximumLaneNumber = 7

var I2CMciMaxSpeedToKhz = map[byte]int{
	0x00: 400,
	0x10: 1_000,
	0x20: 3_400,
	0x30: 0, // reserved onwards
	0x40: 0,
	0x50: 0,
	0x60: 0,
	0x70: 0,
	0x80: 0,
	0x90: 0,
	0xA0: 0,
	0xB0: 0,
	0xC0: 0,
	0xD0: 0,
	0xE0: 0,
	0xF0: 0,
}

var SPIMciMaxSpeedToKhz = map[byte]int{
	0x00: 1_000,
	0x01: 2_000,
	0x02: 4_000,
	0x03: 8_000,
	0x04: 12_000,
	0x05: 16_000,
	0x06: 20_000,
	0x07: 30_000,
	0x08: 40_000,
	0x09: 50_000,
	0x0A: 0, // reserved onwards
	0x0B: 0,
	0x0C: 0,
	0x0D: 0,
	0x0E: 0,
	0x0F: 0,
}

var PowerClassToInt = map[byte]int{
	0b0000_0000: 1,
	0b0010_0000: 2,
	0b0100_0000: 3,
	0b0110_0000: 4,
	0b1000_0000: 5,
	0b1010_0000: 6,
	0b1100_0000: 7,
	0b1110_0000: 8,
}

var MediaInterfaceToStr = map[byte]string{
	0x00: "850nm VCSEL",
	0x01: "1310nm VCSEL",
	0x02: "1550nm VCSEL",
	0x03: "1310nm FP",
	0x04: "1310nm DFB",
	0x05: "1550nm DFB",
	0x06: "1310nm EML",
	0x07: "1550nm EML",
	0x08: "Other",
	0x09: "1490nm DFB",
	0x0A: "Copper cable, passive, unequalized",
	0x0B: "Copper cable, passive, equalized",
	0x0C: "Copper cable with near and far end limiting active equalizers",
	0x0D: "Copper cable with end limiting active equalizers",
	0x0E: "Copper cable with near end limiting active equalizers",
	0x0F: "Copper cable with linear active equalizers",
	0x10: "C-band tunable laser",
	0x11: "L-band tunable laser",
	0x12: "Copper cable with near and far end linear active equalizers",
	0x13: "Copper cable with far end linear active equalizers",
	0x14: "Copper cable with near end linear active equalizers",
}

const (
	BankSelectErrorString                     = "I could not write to bank select register, aborting program now."
	PageSelectErrorString                     = "I could not write to page select register, aborting program now."
	RegisterWriteErrorString                  = "I could not write to arbitrary register, aborting program now."
	PageOrBankNotAvailableTemplateErrorString = "Module responded that page 0x%x with bank 0x%x is not available," +
		" its likely I failed to correctly identify the module capabilities. Aborting program now."
	NoStagedConfigSupportErrorString = "Module has signaled that staged config change is mandatory for any change, " +
		"but I do not have this capability. Aborting program now."
	FlatMemoryMapErrorString = "Module signals a flat memory map. Only lower memory and page 00 are available, " +
		"and you are requesting pages above this. Aborting program now."
	RefusalMajorTooHigh string = "This CMIS module has" +
		" a Major Revision number over what I can speak " +
		"therefore it is unsupported. Program will be terminated " +
		"now, as I cannot read nor write to this module without " +
		"potential failure, data loss and/or equipment damage."
)
