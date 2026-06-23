package cmis

import (
	"fmt"
	"slices"

	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/optic/util"
)

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

const BankSelectErrorString = "I could not write to bank select register, aborting program now."
const PageSelectErrorString = "I could not write to page select register, aborting program now."
const RegisterWriteErrorString = "I could not write to arbitrary register, aborting program now."
const PageOrBankNotAvailableTemplateErrorString = "Module responded that page 0x%x with bank 0x%x is not available," +
	" its likely I failed to correctly identify the module capabilities. Aborting program now."

type ManagementStrategy struct {
	state *pkg.ModuleState
}

type ExtensionManagementStrategy struct {
	state *pkg.ModuleState
}

func New(state *pkg.ModuleState) *ManagementStrategy {
	return &ManagementStrategy{
		state: state,
	}
}

func NewCMISExtension(state *pkg.ModuleState) *ExtensionManagementStrategy {
	return &ExtensionManagementStrategy{
		state: state,
	}
}

func (e *ExtensionManagementStrategy) GetExtensionState() (*pkg.ModuleState, error) {
	_, err := e.GetSupportedControlsAdvertising()
	if err != nil {
		return nil, err
	}
	_, err = e.GetLaserCapabilitiesAdvertising()
	if err != nil {
		return nil, err
	}
	_, err = e.GetTunableLaserControlStatus()
	if err != nil {
		return nil, err
	}

	return e.state, nil
}

func (e *ExtensionManagementStrategy) SetExtensionState(s *pkg.ModuleState) (*pkg.ModuleState, error) {
	//TODO implement me
	panic("implement me")
}

func (e *ExtensionManagementStrategy) ManufacturerIsCompatibleWithProtocolExtension(manufacturer string) bool {
	return true // manufacturer names have 0 influence, only sff8024 does. so we defer decision
}

func (e *ExtensionManagementStrategy) SFF8024IsCompatibleWithProtocolExtension(
	sff8024Identifier byte,
	sff8024Revision byte,
) bool {
	return checkSFF8024(sff8024Identifier, sff8024Revision)
}

func (e *ExtensionManagementStrategy) Activate() (*pkg.ModuleState, error) {
	e.state.CMISOnlyExtension.Active = true
	return e.state, nil
}

func (e *ExtensionManagementStrategy) GetTunableLaserControlStatus() (*pkg.ModuleState, error) {
	// _, err := e.state.GetAdministrativeInformation() should already have been fetched
	if !e.state.CMISOnlyExtension.MemoryModelPaged {
		return e.state, nil // noop
	}

	var bank byte
	for bank = 0x00; bank <= e.state.CMISOnlyExtension.SupportedControls.MaximumBankSupported; bank += 1 {
		e.state.CMISOnlyExtension.TunableLaserCtrlStatus = append(
			e.state.CMISOnlyExtension.TunableLaserCtrlStatus,
			pkg.CMISBankedTunableLaserControlAndStatus{},
		) // adding banks on the go to avoid having max banks all the time
		_, err := GetTunableLaserControlStatus(
			e.state, &e.state.CMISOnlyExtension.TunableLaserCtrlStatus[bank], bank, MaximumLaneNumber,
		)
		if err != nil {
			return nil, err
		}
	}

	return e.state, nil
}

func (e *ExtensionManagementStrategy) GetLaserCapabilitiesAdvertising() (*pkg.ModuleState, error) {
	if !e.state.CMISOnlyExtension.MemoryModelPaged {
		return e.state, nil // noop
	}

	_, err := GetLaserCapabilitiesAdvertising(
		e.state, &e.state.CMISOnlyExtension.LaserCapabilities,
	)
	if err != nil {
		return nil, err
	}

	return e.state, nil
}

func (e *ExtensionManagementStrategy) GetSupportedControlsAdvertising() (*pkg.ModuleState, error) {
	// _, err2 := e.state.GetAdministrativeInformation() should already have been fetched
	// check if page 01 is supported
	if !e.state.CMISOnlyExtension.MemoryModelPaged {
		return e.state, nil // noop
	}

	dumpBin, err := e.state.GetPageBin(0x01, 0x00)
	if err != nil {
		return nil, err
	}

	caps := &e.state.CMISOnlyExtension.SupportedControls

	caps.ModuleInactiveFirmwareMajorRevision = dumpBin[0x80]
	caps.ModuleInactiveFirmwareMinorRevision = dumpBin[0x81]
	caps.ModuleHardwareMajorRevision = dumpBin[0x82]
	caps.ModuleHardwareMinorRevision = dumpBin[0x83]

	// skipping supported lengths

	// only defined for SINGLE WAVELENGTH modules
	var base byte = 0x8A
	caps.NominalWavelengthNm = 0.05 * float64(util.ReadBeUint16AndShiftBase(dumpBin, &base))
	caps.WavelengthToleranceNm = 0.005 * float64(util.ReadBeUint16AndShiftBase(dumpBin, &base))

	// skipping network / vdm / diag / coherent / cmisff / 03h
	caps.MaximumBankSupported = dumpBin[0x8E] & BanksSupportedMask
	if caps.MaximumBankSupported == 0x03 { // signals lane banked
		caps.MaximumBankSupported = dumpBin[0xAE] & ExtraLaneBanksSupportedMask
	}

	// go should have bitfields
	caps.WavelengthIsControllable = dumpBin[0x9B]&WavelengthIsControllableMask != 0
	caps.TransmitterIsTunable = dumpBin[0x9B]&TransmitterIsTunableMask != 0
	caps.SquelchMethodTx = dumpBin[0x9B] & SquelchMethodTxMask
	caps.ForcedSquelchTxSupported = dumpBin[0x9B]&ForcedSquelchTxSupportedMask != 0
	caps.AutoSquelchDisableTxSupported = dumpBin[0x9B]&AutoSquelchDisableTxSupportedMask != 0
	caps.OutputDisableTxSupported = dumpBin[0x9B]&OutputDisableTxSupportedMask != 0
	caps.InputPolarityFlipTxSupported = dumpBin[0x9B]&InputPolarityFlipTxSupportedMask != 0
	caps.BankBroadcastSupported = dumpBin[0x9C]&BankBroadcastSupportedMask != 0
	caps.AutoSquelchDisableRxSupported = dumpBin[0x9C]&AutoSquelchDisableRxSupportedMask != 0
	caps.OutputDisableRxSupported = dumpBin[0x9C]&OutputDisableRxSupportedMask != 0
	caps.OutputPolarityFlipRxSupported = dumpBin[0x9C]&OutputPolarityFlipRxSupportedMask != 0

	return e.state, nil
}

func (s2 *ManagementStrategy) GetPageBin(page byte, bank byte) ([]byte, error) {
	handle := s2.state.GetHandle()

	// we do not have one-shot 2 bytes write available
	// currently this is out of spec
	// > Note: In other words, a host needs to write both the BankSelect and the PageSelect register in a single WRITE
	// > transaction, except for the case when BankIndex value in the BankSelect register does not change. In this case
	// > it is sufficient to WRITE to the PageSelect register.
	// > Note: However, modules may also choose to accept the two registers written in two subsequent WRITE
	// > transactions, to work with non-compliant hosts. Note that rainy day scenarios remain unspecified in this case,
	// > and the PageMapping Validity assertion described below is compromised.

	// do bank select first (will not be cleared if incorrect)
	err := handle.Connection.DoI2CSet(handle.I2cBusId, BankSelectRegisterAddress, bank)
	if err != nil {
		panic(BankSelectErrorString)
	}

	// do page select next (will be cleared if page + bank is incorrect)
	err = handle.Connection.DoI2CSet(handle.I2cBusId, PageSelectRegisterAddress, page)
	if err != nil {
		panic(PageSelectErrorString)
	}

	// decode page
	pageStr, err := handle.Connection.GetI2CDump(handle.I2cBusId)
	if err != nil {
		return []byte{}, err
	}
	dumpBin := pkg.ParseI2CDump(*pageStr)

	// check page + bank select was authorized only and only if target page select = page select
	if dumpBin[PageSelectRegisterAddress] != page {
		panic(fmt.Sprintf(PageOrBankNotAvailableTemplateErrorString, page, bank))
	}

	// return if everything went ok
	return dumpBin, nil
}

func (s2 *ManagementStrategy) WritePageByteBin(page byte, bank byte, offset byte, value byte) error {
	handle := s2.state.GetHandle()

	_, err := s2.GetPageBin(page, bank) // select page by read
	if err != nil {
		return err
	}

	// write is authorized if nothing failed
	err = handle.Connection.DoI2CSet(handle.I2cBusId, (int)(offset), value)
	if err != nil {
		panic(RegisterWriteErrorString)
	}

	return nil
}

const RefusalMajorTooHigh string = "This CMIS module has" +
	" a Major Revision number over what I can speak " +
	"therefore it is unsupported. Program will be terminated " +
	"now, as I cannot read nor write to this module without " +
	"potential failure, data loss and/or equipment damage."

func checkSFF8024(sff8024Identifier byte, sff8024Revision byte) bool {
	var CmisCompatibleSFF8024IDs = [...]byte{
		0x1E, // qsfp+ or later with cmis
		0x1F, // sfp-dd with cmis
		0x20, // sfp+ or later with cmis
		0x21, // osfp-xd with cmis
		0x22, // oif-elfs with cmis
		0x23, // 4 lanes cdfp with cmis
		0x24, // 8 lanes cdfp with cmis
		0x25, // 16 lanes cdfp with cmis
		0x18, // qsfp-dd 8x - may support cmis
	}

	const MajVerMask byte = 0xF0 // upper nibble is bits 7-4 per cmis rev 5.38 section 8.2.1
	const MinMajVer byte = 0x50  // minimum supported major version is 5

	compatibleIdentifier := func(id byte) bool { return slices.Contains(CmisCompatibleSFF8024IDs[:], id) }
	compatibleMajorRev := func(sff8024rev byte) bool { return (sff8024rev & MajVerMask) <= MinMajVer }

	// panic if CMIS and revision is above our maximum
	if compatibleIdentifier(sff8024Identifier) && !compatibleMajorRev(sff8024Revision) {
		panic(RefusalMajorTooHigh)
	}

	// tells if compatible
	return compatibleIdentifier(sff8024Identifier)
}

func (s2 *ManagementStrategy) AcceptsSFF8024(sff8024Identifier byte, sff8024Revision byte) bool {
	return checkSFF8024(sff8024Identifier, sff8024Revision)
}

func (s2 *ManagementStrategy) Set(s *pkg.ModuleState) (*pkg.ModuleState, error) {
	//TODO implement me
	panic("implement me")
}

func (s2 *ManagementStrategy) Get() (*pkg.ModuleState, error) {
	_, err := s2.GetAdministrativeInformation()
	if err != nil {
		return nil, err
	}
	return s2.state, nil
}

func (s2 *ManagementStrategy) GetAdministrativeInformation() (*pkg.ModuleState, error) {
	dumpBin, err := s2.state.GetPageBin(0x00, 0x00)
	if err != nil {
		return nil, err
	}

	s2.state.ManagementProtocol = "cmis"

	// lower mem
	s2.state.CMISOnlyExtension.MemoryModelPaged = dumpBin[0x02]&MemoryModelMask == 0 // 0 is paged memory, 1 is flat
	s2.state.CMISOnlyExtension.SteppedConfigOnly = dumpBin[0x02]&SteppedConfigOnlyMask == 0
	s2.state.CMISOnlyExtension.I2CMciMaxSpeedKhz = I2CMciMaxSpeedToKhz[dumpBin[0x02]&I2CMciMaxSpeedMask]
	s2.state.CMISOnlyExtension.SPIMciMaxSpeedKhz = SPIMciMaxSpeedToKhz[dumpBin[0x02]&SPIMCIMaxSpeedMask]
	AutoCommissioning := dumpBin[0x02] & AutoCommissioningMask
	s2.state.CMISOnlyExtension.AutoCommissioningNone = AutoCommissioning == 0b0000
	s2.state.CMISOnlyExtension.AutoCommissioningRegular = AutoCommissioning == 0b0001
	s2.state.CMISOnlyExtension.AutoCommissioningHot = AutoCommissioning == 0b0010 // 0x11 is reserved

	// page 00
	s2.state.VendorName = util.ParseASCIIToString(dumpBin[0x81:0x90])
	s2.state.CMISOnlyExtension.VendorOUI = dumpBin[0x91:0x93]
	s2.state.VendorPartNumber = util.ParseASCIIToString(dumpBin[0x94:0xA3])
	s2.state.VendorPartRevision = util.ParseASCIIToString(dumpBin[0xA4:0xA5])
	s2.state.VendorSerialNumber = util.ParseASCIIToString(dumpBin[0xA6:0xB5])
	s2.state.CMISOnlyExtension.DateCode = util.ParseASCIIToString(dumpBin[0xB6:0xBD])
	s2.state.CMISOnlyExtension.CLEICode = util.ParseASCIIToString(dumpBin[0xBE:0xC7])
	s2.state.CMISOnlyExtension.PowerClass = PowerClassToInt[dumpBin[0xC8]&ModulePowerClassMask]
	s2.state.CMISOnlyExtension.MaxPowerWatts = 0.25 * float64(dumpBin[0xC9]) // byte is interpreted as uint8. unit is ceil of quarter-watts

	s2.state.CMISOnlyExtension.SupportedMediaLanes = make(map[int]bool)
	// fetching media lane support per lane
	var mask byte = 0b1000_0000
	for i := 8; i >= 1; i -= 1 { // MSB countdown
		s2.state.CMISOnlyExtension.SupportedMediaLanes[i] = dumpBin[0xD2]&mask == 0 // supported == 0, unsupported == 1
		mask >>= 1
	}
	s2.state.CMISOnlyExtension.FarEndDetachableMedia = dumpBin[0xD3]&FarEndConfigurationMask == 0b0000_0000
	s2.state.CMISOnlyExtension.FarEnd1LaneModule = dumpBin[0xD3]&FarEndConfigurationMask == 0b0000_0001
	s2.state.CMISOnlyExtension.FarEnd2LanesModule = dumpBin[0xD3]&FarEndConfigurationMask == 0b000_0110
	s2.state.CMISOnlyExtension.FarEnd4LanesModule = dumpBin[0xD3]&FarEndConfigurationMask == 0b000_0011
	s2.state.CMISOnlyExtension.FarEnd8LanesModule = dumpBin[0xD3]&FarEndConfigurationMask == 0b000_0010
	s2.state.CMISOnlyExtension.FarEnd16LanesModule = dumpBin[0xD3]&FarEndConfigurationMask == 0b001_1011
	s2.state.CMISOnlyExtension.MediaInterface = MediaInterfaceToStr[dumpBin[0xD4]]

	return s2.state, nil
}

func (s2 *ManagementStrategy) SetAdministrativeInformation(s *pkg.ModuleState) (*pkg.ModuleState, error) {
	//TODO implement me
	panic("implement me")
}

// GetTunableLaserControlStatus maxN between 0 and 7 (channel N, used for only-channel-0 access)
func GetTunableLaserControlStatus(state *pkg.ModuleState, caps *pkg.CMISBankedTunableLaserControlAndStatus, bank byte, maxN int) (*pkg.ModuleState, error) {
	dumpBin, err := state.GetPageBin(TunableLaserControlStatusPage, bank)
	if err != nil {
		return nil, err
	}

	for i := 0; i <= maxN; i += 1 {
		caps.GridSpacingTx[i] = dumpBin[0x80+i] & pkg.CMISGridSpacingTxMask
		caps.GridSpacingTxRO[i] = pkg.CMISGridSpacingToFloatGhzMap[caps.GridSpacingTx[i]]
		caps.FineTuningEnableTx[i] = dumpBin[0x80+i]&pkg.CMISFineTuningEnableTxMask != 0

		caps.ChannelNumberTx[i] = util.ReadBeInt16(dumpBin, 0x88+byte(2*i)) // S16 over 2 bytes
		caps.FineTuningOffsetMhzTx[i] = util.ReadBeInt16(dumpBin, 0x98+byte(2*i))
		caps.CurrentLaserFrequencyMhzTx[i] = util.ReadBeUint32(dumpBin, 0xA8+byte(4*i)) // U32 over 4 bytes, units Mhz

		caps.TargetOutputPowerTx[i] = util.ReadBeInt16(dumpBin, 0xC8+byte(2*i))

		caps.TuningInProgressTx[i] = dumpBin[0xDE+i]&pkg.CMISTuningInProgressTxMask != 0
		caps.WaveLengthUnlockStatus[i] = dumpBin[0xDE+i]&pkg.CMISWavelengthUnlockStatusTxMask != 0

		// per spec: the bit n-1 is set if and only if any of the latched flags are set to 1
		caps.LaserTuningFlagSummaryTx[i] = (dumpBin[0xE6+i] & (0b0000_0001 << i)) != 0

		caps.TargetOutputPowerOORFlagTx[i] = dumpBin[0xE7+i]&pkg.CMISTargetOutputPowerOORFlagTxMask != 0
		caps.FineTuningOutOfRangeFlagTx[i] = dumpBin[0xE7+i]&pkg.CMISFineTuningOutOfRangeFlagTxMask != 0
		caps.TuningNotAcceptedMaskTx[i] = dumpBin[0xE7+i]&pkg.CMISTuningNotAcceptedFlagTxMask != 0
		caps.InvalidChannelNumberFLagTx[i] = dumpBin[0xE7+i]&pkg.CMISInvalidChannelNumberFlagTxMask != 0
		caps.WavelengthUnlockedFlagTx[i] = dumpBin[0xE7+i]&pkg.CMISWavelengthUnlockedFlagTxMask != 0
		caps.TuningCompleteFlagTx[i] = dumpBin[0xE7+i]&pkg.CMISTuningCompleteFlagTxMask != 0

		// reading interrupt masks, bitfield pattern same as alarm
		caps.TargetOutputPowerOORMaskTx[i] = dumpBin[0xEF+i]&pkg.CMISTargetOutputPowerOORFlagTxMask != 0
		caps.FineTuningPowerOutOfRangeMaskTx[i] = dumpBin[0xEF+i]&pkg.CMISFineTuningOutOfRangeFlagTxMask != 0
		caps.TuningNotAcceptedMaskTx[i] = dumpBin[0xEF+i]&pkg.CMISTuningNotAcceptedFlagTxMask != 0
		caps.InvalidChannelMaskTx[i] = dumpBin[0xEF+i]&pkg.CMISInvalidChannelNumberFlagTxMask != 0
		caps.WavelengthUnlockedMaskTx[i] = dumpBin[0xEF+i]&pkg.CMISWavelengthUnlockedFlagTxMask != 0
		caps.TuningCompleteMaskTx[i] = dumpBin[0xEF+i]&pkg.CMISTuningCompleteFlagTxMask != 0
	}

	return state, nil
}

func SetTunableLaserControlStatus(state *pkg.ModuleState, caps *pkg.CMISBankedTunableLaserControlAndStatus, bank byte, maxN int) (*pkg.ModuleState, error) {
	dumpBin, err := state.GetPageBin(TunableLaserControlStatusPage, bank)
	if err != nil {
		return nil, err
	}

	for i := 0; i <= maxN; i += 1 {
		// rewrite to mem
		dumpBin[0x80+i] = (caps.GridSpacingTx[i] & pkg.CMISGridSpacingTxMask) &
			(util.YesNoByte(caps.FineTuningEnableTx[i]) & pkg.CMISFineTuningEnableTxMask)

		util.WriteBeInt16(caps.ChannelNumberTx[i], dumpBin, 0x88+byte(2*i))
		util.WriteBeInt16(caps.FineTuningOffsetMhzTx[i], dumpBin, 0x98+byte(2*i))
		// no write for CurrentLaserFrequency, read-only

		util.WriteBeInt16(caps.TargetOutputPowerTx[i], dumpBin, 0xC8+byte(2*i))
		// subsequent fields read-only

		dumpBin[0xEF+i] = (util.YesNoByte(caps.TargetOutputPowerOORMaskTx[i]) & pkg.CMISTargetOutputPowerOORFlagTxMask) &
			(util.YesNoByte(caps.FineTuningPowerOutOfRangeMaskTx[i]) & pkg.CMISFineTuningOutOfRangeFlagTxMask) &
			(util.YesNoByte(caps.TuningNotAcceptedMaskTx[i]) & pkg.CMISTuningNotAcceptedFlagTxMask) &
			(util.YesNoByte(caps.InvalidChannelMaskTx[i]) & pkg.CMISInvalidChannelNumberFlagTxMask) &
			(util.YesNoByte(caps.WavelengthUnlockedMaskTx[i]) & pkg.CMISWavelengthUnlockedFlagTxMask) &
			(util.YesNoByte(caps.TuningCompleteMaskTx[i]) & pkg.CMISTuningCompleteFlagTxMask)
	}

	err = state.WritePageBin(TunableLaserControlStatusPage, bank, dumpBin)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func GetLaserCapabilitiesAdvertising(state *pkg.ModuleState, caps *pkg.CMISLaserCapabilitiesAdvertising) (*pkg.ModuleState, error) {
	dumpBin, err := state.GetPageBin(0x04, 0x00) // non-banked
	if err != nil {
		return nil, err
	}

	caps.SupportedFrequencies = make(map[string]bool)

	// I really hate that Go doesn't have bitfields.
	caps.SupportedFrequencies["75.000"] = dumpBin[0x80]&GridSupported75GhzMask != 0
	caps.SupportedFrequencies["33.000"] = dumpBin[0x80]&GridSupported33GhzMask != 0
	caps.SupportedFrequencies["100.000"] = dumpBin[0x80]&GridSupported100GhzMask != 0
	caps.SupportedFrequencies["50.000"] = dumpBin[0x80]&GridSupported50GhzMask != 0
	caps.SupportedFrequencies["25.000"] = dumpBin[0x80]&GridSupported25GhzMask != 0
	caps.SupportedFrequencies["12.500"] = dumpBin[0x80]&GridSupported12p5GhzMask != 0
	caps.SupportedFrequencies["6.250"] = dumpBin[0x80]&GridSupported6p25GhzMask != 0
	caps.SupportedFrequencies["3.125"] = dumpBin[0x80]&GridSupported3p125GhzMask != 0
	caps.SupportedFrequencies["150.000"] = dumpBin[0x81]&GridSupported150GhzMask != 0

	caps.FineTuningSupported = dumpBin[0x81]&FineTuningSupportedMask != 0

	caps.GridLowChannel = make(map[string]int16)
	caps.GridHighChannel = make(map[string]int16)

	var base byte = 0x82
	// 3.125Ghz
	caps.GridLowChannel["3.125"] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.GridHighChannel["3.125"] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	// 6.25Ghz
	caps.GridLowChannel["6.250"] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.GridHighChannel["6.250"] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	// 12.5Ghz
	caps.GridLowChannel["12.500"] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.GridHighChannel["12.500"] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	// 25Ghz
	caps.GridLowChannel["25.000"] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.GridHighChannel["25.000"] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	// 50Ghz
	caps.GridLowChannel["50.000"] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.GridHighChannel["50.000"] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	// 100Ghz
	caps.GridLowChannel["100.000"] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.GridHighChannel["100.000"] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	// 33Ghz
	caps.GridLowChannel["33.000"] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.GridHighChannel["33.000"] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	// 75Ghz
	caps.GridLowChannel["75.000"] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.GridHighChannel["75.000"] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	// 150Ghz
	caps.GridLowChannel["150.000"] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.GridHighChannel["150.000"] = util.ReadBeInt16AndShiftBase(dumpBin, &base)

	base = 0xBE // skip reserved region
	caps.FineTuningResolution = util.ReadBeUint16AndShiftBase(dumpBin, &base)
	caps.FineTuningLowOffset = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.FineTuningHighOffset = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.ProgOutputPowerPerLaneSupported = dumpBin[base]&ProgOutputPowerPerLaneSupported != 0

	base = 0xC6 // skip reserved region
	caps.ProgOutputPowerMin = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.ProgOutputPowerMax = util.ReadBeInt16AndShiftBase(dumpBin, &base)

	// TODO checksum support
	return state, nil
}
