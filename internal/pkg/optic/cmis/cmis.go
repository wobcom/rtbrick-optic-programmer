package cmis

import (
	"slices"

	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/optic/util"
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

const ProgOutputPowerPerLaneSupported = 0b1000_0000

const TunableLaserControlStatusPage = 0x12

type ManagementStrategy struct {
	state *pkg.ModuleState
}

func New(state *pkg.ModuleState) *ManagementStrategy {
	return &ManagementStrategy{
		state: state,
	}
}

func (s2 *ManagementStrategy) GetPageBin(page byte, bank byte) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s2 *ManagementStrategy) WritePageByteBin(page byte, bank byte, offset byte, value byte) error {
	//TODO implement me
	panic("implement me")
}

const RefusalMajorTooHigh string = "This CMIS module has" +
	" a Major Revision number over what I can speak " +
	"therefore it is unsupported. Program will be terminated " +
	"now, as I cannot read nor write to this module without " +
	"potential failure, data loss and/or equipment damage."

func (s2 *ManagementStrategy) AcceptsSFF8024(sff8024Identifier byte, sff8024Revision byte) bool {
	var CmisCompatibleSFF8024IDs = [...]byte{
		0x1E, // qsfp+ or later with cmis
		0x1F, // sfp-dd with cmis
		0x20, // sfp+ or later with cmis
		0x21, // osfp-xd with cmis
		0x22, // oif-elfs with cmis
		0x23, // 4 lanes cdfp with cmis
		0x24, // 8 lanes cdfp with cmis
		0x25, // 16 lanes cdfp with cmis
	}

	const MajVerMask byte = 0b111_1000 // upper nibble is bits 7-4 per cmis rev 5.38 section 8.2.1
	const MinMajVer byte = 0x05        // minimum supported major version is 5

	compatibleIdentifier := func(id byte) bool { return slices.Contains(CmisCompatibleSFF8024IDs[:], id) }
	compatibleMajorRev := func(sff8024rev byte) bool { return (sff8024Identifier & MajVerMask) <= MinMajVer }

	// panic if CMIS and revision is above our maximum
	if compatibleIdentifier(sff8024Identifier) && !compatibleMajorRev(sff8024Revision) {
		panic(RefusalMajorTooHigh)
	}

	// tells if compatible
	return compatibleIdentifier(sff8024Identifier)
}

func (s2 *ManagementStrategy) Set(s *pkg.ModuleState) (*pkg.ModuleState, error) {
	//TODO implement me
	panic("implement me")
}

func (s2 *ManagementStrategy) Get() (*pkg.ModuleState, error) {
	//TODO implement me
	panic("implement me")
}

func (s2 *ManagementStrategy) GetAdministrativeInformation() (*pkg.ModuleState, error) {
	//TODO implement me
	panic("implement me")
}

func (s2 *ManagementStrategy) SetAdministrativeInformation(s *pkg.ModuleState) (*pkg.ModuleState, error) {
	//TODO implement me
	panic("implement me")
}

func (s2 *ManagementStrategy) GetTunableLaserControlStatus(s *pkg.ModuleState) {
	// TODO implement me
	panic("implement me")
}

func (s2 *ManagementStrategy) GetLaserCapabilitiesAdvertising(s *pkg.ModuleState) (*pkg.ModuleState, error) {
	// TODO implement me
	panic("implement me")
}

// GetTunableLaserControlStatus maxN between 0 and 7 (channel N, used for only-channel-0 access)
func GetTunableLaserControlStatus(s *pkg.ModuleState, bank byte, maxN int) (*pkg.ModuleState, error) {
	dumpBin, err := s.GetPageBin(TunableLaserControlStatusPage, bank)
	if err != nil {
		return nil, err
	}
	caps := &s.FlexOptixSFF8636Extension.TunableLaserCtrlStatus

	for i := 0; i <= maxN; i += 1 {
		caps.GridSpacingTx[i] = dumpBin[0x80+i] & pkg.CMISGridSpacingTxMask
		caps.GridSpacingTxRO[i] = pkg.CMISGridSpacingToFloatGhzMap[caps.GridSpacingTx[i]]
		caps.FineTuningEnableTx[i] = dumpBin[0x80+i]&pkg.CMISFineTuningEnableTxMask != 0

		caps.ChannelNumberTx[i] = util.ReadBeInt16(dumpBin, 0x88+byte(2*i)) // S16 over 2 bytes
		caps.FineTuningOffsetTx[i] = util.ReadBeInt16(dumpBin, 0x98+byte(2*i))
		caps.CurrentLaserFrequencyTx[i] = util.ReadBeUint32(dumpBin, 0xA8+byte(2*i))

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

	return s, nil
}

func SetTunableLaserControlStatus(s *pkg.ModuleState, bank byte, maxN int) (*pkg.ModuleState, error) {
	dumpBin, err := s.GetPageBin(TunableLaserControlStatusPage, bank)
	if err != nil {
		return nil, err
	}
	caps := &s.FlexOptixSFF8636Extension.TunableLaserCtrlStatus

	for i := 0; i <= maxN; i += 1 {
		// rewrite to mem
		dumpBin[0x80+i] = (caps.GridSpacingTx[i] & pkg.CMISGridSpacingTxMask) &
			(util.YesNoByte(caps.FineTuningEnableTx[i]) & pkg.CMISFineTuningEnableTxMask)

		util.WriteBeInt16(caps.ChannelNumberTx[i], dumpBin, 0x88+byte(2*i))
		util.WriteBeInt16(caps.FineTuningOffsetTx[i], dumpBin, 0x98+byte(2*i))
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

	err = s.WritePageBin(TunableLaserControlStatusPage, bank, dumpBin)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func GetLaserCapabilitiesAdvertising(s *pkg.ModuleState) (*pkg.ModuleState, error) {
	dumpBin, err := s.GetPageBin(0x04, 0x00) // non-banked
	if err != nil {
		return nil, err
	}
	caps := &s.FlexOptixSFF8636Extension.LaserCapabilities

	// I really hate that Go doesn't have bitfields.
	caps.GridSupported75Ghz = dumpBin[0x80]&GridSupported75GhzMask != 0
	caps.GridSupported33Ghz = dumpBin[0x80]&GridSupported33GhzMask != 0
	caps.GridSupported100Ghz = dumpBin[0x80]&GridSupported100GhzMask != 0
	caps.GridSupported50Ghz = dumpBin[0x80]&GridSupported50GhzMask != 0
	caps.GridSupported25Ghz = dumpBin[0x80]&GridSupported25GhzMask != 0
	caps.GridSupported12p5Ghz = dumpBin[0x80]&GridSupported12p5GhzMask != 0
	caps.GridSupported6p25Ghz = dumpBin[0x80]&GridSupported6p25GhzMask != 0
	caps.GridSupported3p125Ghz = dumpBin[0x80]&GridSupported3p125GhzMask != 0

	caps.FineTuningSupported = dumpBin[0x81]&FineTuningSupportedMask != 0
	caps.GridSupported150Ghz = dumpBin[0x81]&GridSupported150GhzMask != 0

	var base byte = 0x82
	// 3.125Ghz
	caps.GridLowChannel3p125Ghz = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.GridHighChannel3p125Ghz = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	// 6.25Ghz
	caps.GridLowChannel6p25Ghz = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.GridHighChannel6p25Ghz = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	// 12.5Ghz
	caps.GridLowChannel12p5Ghz = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.GridHighChannel12p5Ghz = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	// 25Ghz
	caps.GridLowChannel25Ghz = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.GridHighChannel25Ghz = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	// 50Ghz
	caps.GridLowChannel50Ghz = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.GridHighChannel50Ghz = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	// 100Ghz
	caps.GridLowChannel100Ghz = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.GridHighChannel100Ghz = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	// 33Ghz
	caps.GridLowChannel33Ghz = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.GridHighChannel33Ghz = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	// 75Ghz
	caps.GridLowChannel75Ghz = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.GridHighChannel75Ghz = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	// 150Ghz
	caps.GridLowChannel150Ghz = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.GridHighChannel150Ghz = util.ReadBeInt16AndShiftBase(dumpBin, &base)

	base = 0xBE // skip reserved region
	caps.FineTuningResolution = util.ReadBeUint16AndShiftBase(dumpBin, &base)
	caps.FineTuningLowOffset = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.FineTuningHighOffset = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.ProgOutputPowerPerLaneSupported = dumpBin[base]&ProgOutputPowerPerLaneSupported != 0

	base = 0xC6 // skip reserved region
	caps.ProgOutputPowerMin = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.ProgOutputPowerMax = util.ReadBeInt16AndShiftBase(dumpBin, &base)

	// TODO checksum support
	return s, nil
}
