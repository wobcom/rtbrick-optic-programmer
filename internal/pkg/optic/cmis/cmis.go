package cmis

import (
	"slices"

	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg"
)

type State pkg.ModuleState

const RefusalMajorTooHigh string = "This CMIS module has" +
	" a Major Revision number over what I can speak " +
	"therefore it is unsupported. Program will be terminated " +
	"now, as I cannot read nor write to this module without " +
	"potential failure, data loss and/or equipment damage."

func (s2 *State) AcceptsSFF8024(sff8024Identifier byte, sff8024Revision byte) bool {
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
