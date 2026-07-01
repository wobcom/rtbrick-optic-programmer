package sff8636

import (
	"regexp"

	"github.com/wobcom/rtbrick-optic-programmer/internal/optic"
	"github.com/wobcom/rtbrick-optic-programmer/internal/optic/cmis"
)

type FlexOptixSFF8636ManagementStrategy struct {
	state *optic.ModuleState
}

var ValidFlexOptixVendorName = regexp.MustCompile(`(?i)^flexoptix$`)

func NewFlexOptixSFF8636Extension(state *optic.ModuleState) *FlexOptixSFF8636ManagementStrategy {
	return &FlexOptixSFF8636ManagementStrategy{
		state: state,
	}
}

func (f *FlexOptixSFF8636ManagementStrategy) GetExtensionState() (*optic.ModuleState, error) {
	_, err := cmis.GetLaserCapabilitiesAdvertising(f.state, &f.state.FlexOptixSFF8636.TunableLaser.Capabilities)
	if err != nil {
		return nil, err
	}

	f.state.FlexOptixSFF8636.TunableLaser.CtrlStatus = nil
	f.state.FlexOptixSFF8636.TunableLaser.CtrlStatus = append(
		f.state.FlexOptixSFF8636.TunableLaser.CtrlStatus,
		optic.CMISBankedTunableLaserControlAndStatus{},
	)
	_, err = cmis.GetTunableLaserControlStatus(
		f.state, &f.state.FlexOptixSFF8636.TunableLaser.CtrlStatus[0], 0x00, 0,
	)
	if err != nil {
		return nil, err
	}

	return f.state, nil
}

func (f *FlexOptixSFF8636ManagementStrategy) SetExtensionState() (*optic.ModuleState, error) {
	// laser capabilities advertising is all-fields read-only so, no-op for this one

	_, err := cmis.SetTunableLaserControlStatus(
		f.state, &f.state.FlexOptixSFF8636.TunableLaser.CtrlStatus[0], 0x00, 0,
	)
	if err != nil {
		return nil, err
	}

	return f.state, nil
}

func (f *FlexOptixSFF8636ManagementStrategy) ManufacturerIsCompatibleWithProtocolExtension(manufacturer string) bool {
	return ValidFlexOptixVendorName.MatchString(manufacturer)
}

func (f *FlexOptixSFF8636ManagementStrategy) SFF8024IsCompatibleWithProtocolExtension(sff8024Identifier byte, sff8024Revision byte) bool {
	return checkSFF8024(sff8024Identifier, sff8024Revision)
}

func (f *FlexOptixSFF8636ManagementStrategy) Activate() (*optic.ModuleState, error) {
	f.state.FlexOptixSFF8636.Active = true
	return f.state, nil
}
