package sff8636

const PageSelectRegisterAddress = 0x7F

const (
	PageSelectRegisterWriteErrorString  = "I could not write to Page Select register, aborting program now."
	RegisterWriteErrorString            = "I could not write to arbitrary register, aborting program now."
	PageNotAvailableTemplateErrorString = "Module responded that page 0x%x is not available, its likely I failed to " +
		"correctly identify the module capabilities. Aborting program now."
	RefusalVerMismatch = "This module has an " +
		"SFF8636 Revision number that I do not know of, therefore it is unsupported. Program will be terminated now, " +
		"as I cannot read nor write to this module without potential failure, data loss and/or equipment damage."
)

// page 00 register 0x5D
const (
	SoftwareResetMask          = 0b1000_0000
	EnableHighPowerClass8Mask  = 0b0000_1000
	EnableHighPowerClass57Mask = 0b0000_0100
	LowPwrRequestSWMask        = 0b0000_0010
	LowPwrOverrideMask         = 0b0000_0001
)
