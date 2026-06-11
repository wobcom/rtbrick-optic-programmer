package sff8636

import (
	"slices"

	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/optic/util"
)

// ManagementStrategy SFF8636 is a concrete implementation of the Management interface for SFF8636 spec
type ManagementStrategy struct {
	state *pkg.ModuleState
}

type ExtensionManagementStrategy struct {
	state *pkg.ModuleState
}

const PageSelectRegisterWriteErrorString = "I could not write to Page Select register, aborting program now."
const PageNotAvailableErrorString = "Module responded that this page is not available, its likely I failed to " +
	"correctly identify the module capabilities. Aborting program now."

func NewSFF8636Extension(state *pkg.ModuleState) *ExtensionManagementStrategy {
	return &ExtensionManagementStrategy{
		state: state,
	}
}

func (s2 ManagementStrategy) GetPageBin(page byte, _ byte) ([]byte, error) {
	const PageSelectRegisterAddress = 0x7F
	handle := s2.state.GetHandle()

	// do raw Page Select write 1st
	pageSelectErr := handle.Connection.DoI2CSet(handle.I2cBusId, PageSelectRegisterAddress, page)
	if pageSelectErr != nil {
		panic(PageSelectRegisterWriteErrorString)
	}

	// then get page and decode
	pageStr, err := handle.Connection.GetI2CDump(handle.I2cBusId)
	if err != nil {
		return []byte{}, err
	}
	dumpBin := pkg.ParseI2CDump(*pageStr)

	// then check Page Select was authorized by reading back Page Select register
	if dumpBin[PageSelectRegisterAddress] != page {
		panic(PageNotAvailableErrorString)
	}

	// return if everything went O.K.
	return dumpBin, nil
}

func (s2 ManagementStrategy) WritePageBin(page byte, bank byte, offset byte, value byte) error {
	//TODO implement me
	panic("implement me")
}

func (s2 ExtensionManagementStrategy) GetExtensionState() (*pkg.ModuleState, error) {
	// implement SFF8636 specific pages here if needed in the future
	return s2.state, nil
}

func (s2 ExtensionManagementStrategy) SetExtensionState(s *pkg.ModuleState) (*pkg.ModuleState, error) {
	// implement SFF8636 specific pages here if needed in the future
	return s2.state, nil
}

func (s2 ExtensionManagementStrategy) ManufacturerIsCompatibleWithProtocolExtension(_ string) bool {
	return true // manufacturer names have 0 influence, only sff8024 does. so we defer decision
}

func (s2 ExtensionManagementStrategy) SFF8024IsCompatibleWithProtocolExtension(
	sff8024Identifier byte,
	sff8024Revision byte,
) bool {
	return checkSFF8024(sff8024Identifier, sff8024Revision)
}

func (s2 ExtensionManagementStrategy) Activate() (*pkg.ModuleState, error) {
	s2.state.SFF8636OnlyExtension.Active = true
	return s2.state, nil
}

func New(state *pkg.ModuleState) *ManagementStrategy {
	return &ManagementStrategy{
		state: state,
	}
}

const RefusalVerMismatch = "This module has an " +
	"SFF8636 Revision number that I do not know of, therefore it is unsupported. Program will be terminated now, " +
	"as I cannot read nor write to this module without potential failure, data loss and/or equipment damage."

func (s2 ManagementStrategy) AcceptsSFF8024(sff8024Identifier byte, sff8024Revision byte) bool {
	return checkSFF8024(sff8024Identifier, sff8024Revision)
}

func checkSFF8024(sff8024Identifier byte, sff8024Revision byte) bool {
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

func (s2 ManagementStrategy) Set(s *pkg.ModuleState) (*pkg.ModuleState, error) {
	//TODO implement me
	panic("implement me")
}

func (s2 ManagementStrategy) Get() (*pkg.ModuleState, error) {
	_, err := s2.GetAdministrativeInformation()
	if err != nil {
		return nil, err
	}
	// TODO call the rest of getters
	return s2.state, nil
}

func (s2 ManagementStrategy) GetAdministrativeInformation() (*pkg.ModuleState, error) {
	// register 0x5D lower mem masks
	const SoftwareResetMask = 0b1000_0000
	const EnableHighPowerClass8Mask = 0b0000_1000
	const EnableHighPowerClass57Mask = 0b0000_0100
	const LowPwrRequestSWMask = 0b0000_0010
	const LowPwrOverrideMask = 0b0000_0001

	bin, err := s2.state.GetPageBin(0x00, 0)
	if err != nil {
		return nil, err
	}

	s2.state.ManagementProtocol = "sff8636"
	// lower mem
	s2.state.SFF8024Identifier = bin[0x00]
	s2.state.SFF8024Revision = bin[0x01]
	s2.state.SoftwareReset = bin[0x5D]&SoftwareResetMask == SoftwareResetMask
	s2.state.EnableHighPowerClass8 = bin[0x5D]&EnableHighPowerClass8Mask == EnableHighPowerClass8Mask
	s2.state.EnableHighPowerClass57 = bin[0x5D]&EnableHighPowerClass57Mask == EnableHighPowerClass57Mask
	s2.state.LowPwrRequestSW = bin[0x5D]&LowPwrRequestSWMask == LowPwrRequestSWMask
	s2.state.LowPwrOverride = bin[0x5D]&LowPwrOverrideMask == LowPwrOverrideMask

	// page 0x00
	s2.state.VendorName = util.ParseASCIIToString(bin[0x94:0xA3])
	s2.state.VendorPartNumber = util.ParseASCIIToString(bin[0xA8:0xB7])
	s2.state.VendorPartRevision = util.ParseASCIIToString(bin[0xA8:0xB7])
	s2.state.VendorSerialNumber = util.ParseASCIIToString(bin[0xC4:0xD3])

	return s2.state, nil
}

func (s2 ManagementStrategy) SetAdministrativeInformation(s *pkg.ModuleState) (*pkg.ModuleState, error) {
	//TODO implement me
	panic("implement me")
}
