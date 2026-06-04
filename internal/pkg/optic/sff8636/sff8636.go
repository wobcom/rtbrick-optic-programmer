package sff8636

import (
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg"
)

// State SFF8636 is a concrete implementation of the Management interface for SFF8636 spec
type State pkg.ModuleState

func (s2 *State) Accepts(sff8024Identifier byte) bool {
	return sff8024Identifier == 0x18
}

func (s2 *State) Set(s *pkg.ModuleState) (*pkg.ModuleState, error) {
	//TODO implement me
	panic("implement me")
}

func (s2 *State) Get() (*pkg.ModuleState, error) {
	//TODO implement me
	panic("implement me")
}

func (s2 *State) GetAdministrativeInformation() (*pkg.ModuleState, error) {
	//TODO implement me
	panic("implement me")
}

func (s2 *State) SetAdministrativeInformation(s *pkg.ModuleState) (*pkg.ModuleState, error) {
	//TODO implement me
	panic("implement me")
}

func (s2 *State) GetTunableLaserCtrlStatus() (*pkg.ModuleState, error) {
	//TODO implement me
	panic("implement me")
}

func (s2 *State) SetTunableLaserCtrlStatus(s *pkg.ModuleState) (*pkg.ModuleState, error) {
	//TODO implement me
	panic("implement me")
}
