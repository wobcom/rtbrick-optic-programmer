package optic

import (
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg"
)

// State is the default Management strategy which should be set at initialization
// it can only read page00, and does 0 write. It can tell which concrete strategy which you should
// be using by looking at identifier codes and constructor ASCII
type State pkg.ModuleState

func (d *State) Accepts(_ byte) bool {
	return true
}

func (d *State) Set(s *pkg.ModuleState) (*pkg.ModuleState, error) {
	//TODO implement me
	panic(genericWriteErrorString)
}

func (d *State) Get() (*pkg.ModuleState, error) {
	//TODO implement me
	panic(genericReadErrorString)
}

func (d *State) GetAdministrativeInformation() (*pkg.ModuleState, error) {
	// TODO
	panic("implement me")
}

func (d *State) SetAdministrativeInformation(s *pkg.ModuleState) (*pkg.ModuleState, error) {
	panic(genericWriteErrorString)
}

func (d *State) GetTunableLaserCtrlStatus() (*pkg.ModuleState, error) {
	panic(genericReadErrorString)
}

func (d *State) SetTunableLaserCtrlStatus(s *pkg.ModuleState) (*pkg.ModuleState, error) {
	panic(genericWriteErrorString)
}

var genericWriteErrorString = "Module type is unknown, therefore I cannot go on, " +
	"and refuse to set any value as it might have undesired effect. " +
	"Program will be terminated now, as it is likely I failed to detect the Module Type."

var genericReadErrorString = "Module type is unknown, therefore I cannot go on, " +
	"as you require me making sense of things I do not know the shape, " +
	"nor give an absolute and complete answer, for I do not understand all that is before me. " +
	"Program will be terminated now, as it is likely I failed to detect the Module Type."
