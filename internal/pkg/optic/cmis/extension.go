package cmis

import (
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/optic/util"
)

type ExtensionManagementStrategy struct {
	state *pkg.ModuleState
}

func NewCMISExtension(state *pkg.ModuleState) *ExtensionManagementStrategy {
	return &ExtensionManagementStrategy{
		state: state,
	}
}

func (e *ExtensionManagementStrategy) assumeHigherPagesReadable() {
	if !e.state.CMIS.MemoryModelPaged {
		panic(FlatMemoryMapErrorString)
	}
}

// assumeConfigChange ensures that module is placed in a config change state prior to write attempt.
// Requires that administrative information has already been obtained prior to write attempt
func (e *ExtensionManagementStrategy) assumeConfigChange() {
	e.assumeHigherPagesReadable()
	if e.state.CMIS.SteppedConfigOnly {
		panic(NoStagedConfigSupportErrorString)
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
	e.assumeConfigChange()

	_, err := e.SetTunableLaserControlStatus(e.state)
	if err != nil {
		return nil, err
	}

	return e.state, nil
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
	e.state.CMIS.Active = true
	return e.state, nil
}

func (e *ExtensionManagementStrategy) GetTunableLaserControlStatus() (*pkg.ModuleState, error) {
	e.assumeHigherPagesReadable()

	var bank byte
	e.state.CMIS.TunableLaser.CtrlStatus = nil
	for bank = 0x00; bank <= e.state.CMIS.SupportedControls.MaximumBankSupported; bank += 1 {
		e.state.CMIS.TunableLaser.CtrlStatus = append(
			e.state.CMIS.TunableLaser.CtrlStatus,
			pkg.CMISBankedTunableLaserControlAndStatus{},
		) // adding banks on the go to avoid having max banks all the time
		_, err := GetTunableLaserControlStatus(
			e.state, &e.state.CMIS.TunableLaser.CtrlStatus[bank], bank, MaximumLaneNumber,
		)
		if err != nil {
			return nil, err
		}
	}

	return e.state, nil
}

func (e *ExtensionManagementStrategy) SetTunableLaserControlStatus(s *pkg.ModuleState) (*pkg.ModuleState, error) {
	e.assumeHigherPagesReadable()

	for i, bank := range e.state.CMIS.TunableLaser.CtrlStatus {
		_, err := SetTunableLaserControlStatus(e.state, &bank, byte(i), MaximumLaneNumber)
		if err != nil {
			return nil, err
		}
	}

	return e.state, nil
}

func (e *ExtensionManagementStrategy) GetLaserCapabilitiesAdvertising() (*pkg.ModuleState, error) {
	e.assumeHigherPagesReadable()

	_, err := GetLaserCapabilitiesAdvertising(
		e.state, &e.state.CMIS.TunableLaser.Capabilities,
	)
	if err != nil {
		return nil, err
	}

	return e.state, nil
}

func (e *ExtensionManagementStrategy) GetSupportedControlsAdvertising() (*pkg.ModuleState, error) {
	e.assumeHigherPagesReadable()

	dumpBin, err := e.state.GetPageBin(0x01, 0x00)
	if err != nil {
		return nil, err
	}

	caps := &e.state.CMIS.SupportedControls

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
