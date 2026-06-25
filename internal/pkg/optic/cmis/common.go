package cmis

import (
	"slices"

	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/optic/util"
)

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

// GetTunableLaserControlStatus maxN between 0 and 7 (channel N, used for only-channel-0 access)
func GetTunableLaserControlStatus(
	state *pkg.ModuleState,
	caps *pkg.CMISBankedTunableLaserControlAndStatus,
	bank byte,
	maxN int,
) (*pkg.ModuleState, error) {
	dumpBin, err := state.GetPageBin(TunableLaserControlStatusPage, bank)
	if err != nil {
		return nil, err
	}

	for i := 0; i <= maxN; i += 1 {
		caps.GridSpacingTx[i] = dumpBin[0x80+i] & pkg.CMISGridSpacingTxMask
		caps.GridSpacingTxROGhz[i] = pkg.CMISGridSpacingToFloatGhzMap[caps.GridSpacingTx[i]]
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
		caps.InvalidChannelNumberFlagTx[i] = dumpBin[0xE7+i]&pkg.CMISInvalidChannelNumberFlagTxMask != 0
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

func SetTunableLaserControlStatus(
	state *pkg.ModuleState,
	caps *pkg.CMISBankedTunableLaserControlAndStatus,
	bank byte,
	maxN int,
) (*pkg.ModuleState, error) {
	dumpBin, err := state.GetPageBin(TunableLaserControlStatusPage, bank)
	if err != nil {
		return nil, err
	}

	for i := 0; i <= maxN; i += 1 {
		// rewrite to mem
		dumpBin[0x80+i] = (caps.GridSpacingTx[i] & pkg.CMISGridSpacingTxMask) |
			(util.YesNoByte(caps.FineTuningEnableTx[i]) & pkg.CMISFineTuningEnableTxMask)

		util.WriteBeInt16(caps.ChannelNumberTx[i], dumpBin, 0x88+byte(2*i))
		util.WriteBeInt16(caps.FineTuningOffsetMhzTx[i], dumpBin, 0x98+byte(2*i))
		// no write for CurrentLaserFrequency, read-only

		util.WriteBeInt16(caps.TargetOutputPowerTx[i], dumpBin, 0xC8+byte(2*i))
		// subsequent fields read-only

		dumpBin[0xEF+i] = (util.YesNoByte(caps.TargetOutputPowerOORMaskTx[i]) & pkg.CMISTargetOutputPowerOORFlagTxMask) |
			(util.YesNoByte(caps.FineTuningPowerOutOfRangeMaskTx[i]) & pkg.CMISFineTuningOutOfRangeFlagTxMask) |
			(util.YesNoByte(caps.TuningNotAcceptedMaskTx[i]) & pkg.CMISTuningNotAcceptedFlagTxMask) |
			(util.YesNoByte(caps.InvalidChannelMaskTx[i]) & pkg.CMISInvalidChannelNumberFlagTxMask) |
			(util.YesNoByte(caps.WavelengthUnlockedMaskTx[i]) & pkg.CMISWavelengthUnlockedFlagTxMask) |
			(util.YesNoByte(caps.TuningCompleteMaskTx[i]) & pkg.CMISTuningCompleteFlagTxMask)
	}

	err = state.WritePageBin(TunableLaserControlStatusPage, bank, dumpBin)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func GetLaserCapabilitiesAdvertising(
	state *pkg.ModuleState,
	caps *pkg.CMISLaserCapabilitiesAdvertising,
) (*pkg.ModuleState, error) {
	dumpBin, err := state.GetPageBin(0x04, 0x00) // non-banked
	if err != nil {
		return nil, err
	}

	caps.SupportedGridSpacings = make(map[string]bool)

	// I really hate that Go doesn't have bitfields.
	caps.SupportedGridSpacings[pkg.Grid75string] = dumpBin[0x80]&GridSupported75GhzMask != 0
	caps.SupportedGridSpacings[pkg.Grid33String] = dumpBin[0x80]&GridSupported33GhzMask != 0
	caps.SupportedGridSpacings[pkg.Grid100String] = dumpBin[0x80]&GridSupported100GhzMask != 0
	caps.SupportedGridSpacings[pkg.Grid50String] = dumpBin[0x80]&GridSupported50GhzMask != 0
	caps.SupportedGridSpacings[pkg.Grid25String] = dumpBin[0x80]&GridSupported25GhzMask != 0
	caps.SupportedGridSpacings[pkg.Grid12p5String] = dumpBin[0x80]&GridSupported12p5GhzMask != 0
	caps.SupportedGridSpacings[pkg.Grid6p250String] = dumpBin[0x80]&GridSupported6p25GhzMask != 0
	caps.SupportedGridSpacings[pkg.Grid3p125String] = dumpBin[0x80]&GridSupported3p125GhzMask != 0
	caps.SupportedGridSpacings[pkg.Grid150String] = dumpBin[0x81]&GridSupported150GhzMask != 0

	caps.FineTuningSupported = dumpBin[0x81]&FineTuningSupportedMask != 0

	caps.GridLowChannel = make(map[string]int16)
	caps.GridHighChannel = make(map[string]int16)

	var base byte = 0x82
	// 3.125Ghz
	caps.GridLowChannel[pkg.Grid3p125String] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.GridHighChannel[pkg.Grid3p125String] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	// 6.25Ghz
	caps.GridLowChannel[pkg.Grid6p250String] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.GridHighChannel[pkg.Grid6p250String] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	// 12.5Ghz
	caps.GridLowChannel[pkg.Grid12p5String] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.GridHighChannel[pkg.Grid12p5String] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	// 25Ghz
	caps.GridLowChannel[pkg.Grid25String] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.GridHighChannel[pkg.Grid25String] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	// 50Ghz
	caps.GridLowChannel[pkg.Grid50String] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.GridHighChannel[pkg.Grid50String] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	// 100Ghz
	caps.GridLowChannel[pkg.Grid100String] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.GridHighChannel[pkg.Grid100String] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	// 33Ghz
	caps.GridLowChannel[pkg.Grid33String] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.GridHighChannel[pkg.Grid33String] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	// 75Ghz
	caps.GridLowChannel[pkg.Grid75string] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.GridHighChannel[pkg.Grid75string] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	// 150Ghz
	caps.GridLowChannel[pkg.Grid150String] = util.ReadBeInt16AndShiftBase(dumpBin, &base)
	caps.GridHighChannel[pkg.Grid150String] = util.ReadBeInt16AndShiftBase(dumpBin, &base)

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
