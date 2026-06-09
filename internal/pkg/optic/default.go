package optic

import (
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg"
)

// State is the default Management strategy which should be set at initialization
// it can only read page00, and does 0 write. It can tell which concrete strategy which you should
// be using by looking at identifier codes and constructor ASCII
type State pkg.ModuleStateWithDirectPageAccess

func (d *State) AcceptsSFF8024(_ byte, _ byte) bool {
	return true
}

func (d *State) Set(_ *pkg.ModuleState) (*pkg.ModuleState, error) {
	panic(genericWriteErrorString)
}

func (d *State) Get() (*pkg.ModuleState, error) {
	panic(genericReadErrorString)
}

func (d *State) GetAdministrativeInformation() (*pkg.ModuleState, error) {
	bin, err := d.GetPageBin(0x00)
	if err != nil {
		return nil, err
	}

	d.SFF8024Identifier = bin[0x00]
	d.SFF8024Revision = bin[0x01]

	return &d.ModuleState, nil
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
