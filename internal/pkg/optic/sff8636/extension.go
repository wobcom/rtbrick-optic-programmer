package sff8636

import "github.com/wobcom/rtbrick-optic-programmer/internal/pkg"

type ExtensionManagementStrategy struct {
	state *pkg.ModuleState
}

func NewSFF8636Extension(state *pkg.ModuleState) *ExtensionManagementStrategy {
	return &ExtensionManagementStrategy{
		state: state,
	}
}

func (s2 *ExtensionManagementStrategy) GetExtensionState() (*pkg.ModuleState, error) {
	// implement SFF8636 specific pages here if needed in the future
	return s2.state, nil
}

func (s2 *ExtensionManagementStrategy) SetExtensionState(s *pkg.ModuleState) (*pkg.ModuleState, error) {
	// implement SFF8636 specific pages here if needed in the future
	return s2.state, nil
}

func (s2 *ExtensionManagementStrategy) ManufacturerIsCompatibleWithProtocolExtension(_ string) bool {
	return true // manufacturer names have 0 influence, only sff8024 does. so we defer decision
}

func (s2 *ExtensionManagementStrategy) SFF8024IsCompatibleWithProtocolExtension(
	sff8024Identifier byte,
	sff8024Revision byte,
) bool {
	return checkSFF8024(sff8024Identifier, sff8024Revision)
}

func (s2 *ExtensionManagementStrategy) Activate() (*pkg.ModuleState, error) {
	s2.state.SFF8636.Active = true
	return s2.state, nil
}
