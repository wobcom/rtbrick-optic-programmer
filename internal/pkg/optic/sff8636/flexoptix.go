package sff8636

import (
	"regexp"

	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/optic/cmis"
)

type FlexOptixSFF8636ManagementStrategy struct {
	state *pkg.ModuleState
}

var ValidFlexOptixVendorName = regexp.MustCompile(`(?i)^flexoptix$`)

func NewFlexOptixSFF8636Extension(state *pkg.ModuleState) *FlexOptixSFF8636ManagementStrategy {
	return &FlexOptixSFF8636ManagementStrategy{
		state: state,
	}
}

func (f *FlexOptixSFF8636ManagementStrategy) GetExtensionState() (*pkg.ModuleState, error) {
	_, err := cmis.GetLaserCapabilitiesAdvertising(f.state, &f.state.FlexOptixSFF8636Extension.LaserCapabilities)
	if err != nil {
		return nil, err
	}

	_, err = cmis.GetTunableLaserControlStatus(
		f.state, &f.state.FlexOptixSFF8636Extension.TunableLaserCtrlStatus, 0x00, 0,
	)
	if err != nil {
		return nil, err
	}

	return f.state, nil
}

func (f *FlexOptixSFF8636ManagementStrategy) SetExtensionState(s *pkg.ModuleState) (*pkg.ModuleState, error) {
	// laser capabilities advertising is all-fields read-only so, no-op for this one

	_, err := cmis.SetTunableLaserControlStatus(
		f.state, &f.state.FlexOptixSFF8636Extension.TunableLaserCtrlStatus, 0x00, 0,
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

func (f *FlexOptixSFF8636ManagementStrategy) Activate() (*pkg.ModuleState, error) {
	f.state.FlexOptixSFF8636Extension.Active = true
	return f.state, nil
}
