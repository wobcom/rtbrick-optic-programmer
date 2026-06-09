package sff8636

import (
	"slices"

	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/optic/util"
)

// State SFF8636 is a concrete implementation of the Management interface for SFF8636 spec
type State pkg.ModuleStateWithDirectPageAccess

const RefusalVerMismatch = "This module has an " +
	"SFF8636 Revision number that I do not know of, therefore it is unsupported. Program will be terminated now, " +
	"as I cannot read nor write to this module without potential failure, data loss and/or equipment damage."

func (s2 *State) AcceptsSFF8024(sff8024Identifier byte, sff8024Revision byte) bool {
	var SFF8636CompatibleSFF8024IDs = [...]byte{
		0x0D, // qsfp+ or later with sff8646 or sff8436 mgmt interface
		0x11, // qsfp28 or later with sff-8636 management interface
	}

	var CompatibleSFF8638Versions = [...]byte{
		// 0x00 not compatible, do not use for rev 2.5 or higher
		// 0x01 sff 8436 not compatible
		// 0x02 sff 8436 not compatible
		0x03, // 1.3 or earlier
		0x04, // 1.4
		0x05, // 1.5
		0x06, // 2.0
		0x07, // 2.5, 2.6 and 2.7
		0x08, // 2.8, 2.9 and 2.10
		0x09, // 2.11
		0x0A, // 2.12
	}

	compatibleIdentifier := func(id byte) bool { return slices.Contains(SFF8636CompatibleSFF8024IDs[:], id) }
	compatibleRev := func(sff8024rev byte) bool { return slices.Contains(CompatibleSFF8638Versions[:], sff8024rev) }

	// panic if revision is unknown
	if compatibleIdentifier(sff8024Identifier) && !compatibleRev(sff8024Revision) {
		panic(RefusalVerMismatch)
	}

	// tells if compatible
	return compatibleIdentifier(sff8024Identifier)
}

func (s2 *State) Set(s *pkg.ModuleState) (*pkg.ModuleState, error) {
	//TODO implement me
	panic("implement me")
}

func (s2 *State) Get() (*pkg.ModuleState, error) {
	_, err := s2.GetAdministrativeInformation()
	if err != nil {
		return nil, err
	}

	//TODO implement me
	panic("implement me")

	return &s2.ModuleState, nil
}

func (s2 *State) GetAdministrativeInformation() (*pkg.ModuleState, error) {
	bin, err := s2.GetPageBin(0x00)
	if err != nil {
		return nil, err
	}

	s2.ManagementProtocol = "sff8636"
	// lower mem
	s2.SFF8024Identifier = bin[0x00]
	s2.SFF8024Revision = bin[0x01]
	// page 0x00
	s2.VendorName = util.ParseASCIIToString(bin[0x94:0xA3])
	s2.VendorPartNumber = util.ParseASCIIToString(bin[0xA8:0xB7])
	s2.VendorPartRevision = util.ParseASCIIToString(bin[0xA8:0xB7])
	s2.VendorSerialNumber = util.ParseASCIIToString(bin[0xC4:0xD3])

	return &s2.ModuleState, nil
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
