package sff8636

import (
	"fmt"

	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/optic/util"
)

// ManagementStrategy SFF8636 is a concrete implementation of the Management interface for SFF8636 spec
type ManagementStrategy struct {
	state *pkg.ModuleState
}

func New(state *pkg.ModuleState) *ManagementStrategy {
	return &ManagementStrategy{
		state: state,
	}
}

func (s2 *ManagementStrategy) GetPageBin(page byte, _ byte) ([]byte, error) {
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
		panic(fmt.Sprintf(PageNotAvailableTemplateErrorString, page))
	}

	// return if everything went O.K.
	return dumpBin, nil
}

func (s2 *ManagementStrategy) WritePageByteBin(page byte, _ byte, offset byte, value byte) error {
	handle := s2.state.GetHandle()

	// do raw Page Select Write 1st
	pageSelectErr := handle.Connection.DoI2CSet(handle.I2cBusId, PageSelectRegisterAddress, page)
	if pageSelectErr != nil {
		panic(PageSelectRegisterWriteErrorString)
	}

	err := handle.Connection.DoI2CSet(handle.I2cBusId, (int)(offset), value)
	if err != nil {
		panic(RegisterWriteErrorString)
	}

	return nil
}

func (s2 *ManagementStrategy) AcceptsSFF8024(sff8024Identifier byte, sff8024Revision byte) bool {
	return checkSFF8024(sff8024Identifier, sff8024Revision)
}

func (s2 *ManagementStrategy) Set() (*pkg.ModuleState, error) {
	_, err := s2.SetAdministrativeInformation()
	if err != nil {
		return nil, err
	}
	return s2.state, nil // noop
}

func (s2 *ManagementStrategy) Get() (*pkg.ModuleState, error) {
	_, err := s2.GetAdministrativeInformation()
	if err != nil {
		return nil, err
	}
	return s2.state, nil
}

func (s2 *ManagementStrategy) GetAdministrativeInformation() (*pkg.ModuleState, error) {
	bin, err := s2.state.GetPageBin(0x00, 0)
	if err != nil {
		return nil, err
	}

	s2.state.ManagementProtocol = "sff8636"
	// lower mem
	s2.state.SFF8024Identifier = bin[0x00]
	s2.state.SFF8024Revision = bin[0x01]
	s2.state.SoftwareReset = bin[0x5D]&SoftwareResetMask == SoftwareResetMask
	s2.state.SFF8636.EnableHighPowerClass8 = bin[0x5D]&EnableHighPowerClass8Mask == EnableHighPowerClass8Mask
	s2.state.SFF8636.EnableHighPowerClass57 = bin[0x5D]&EnableHighPowerClass57Mask == EnableHighPowerClass57Mask
	s2.state.LowPwrRequestSW = bin[0x5D]&LowPwrRequestSWMask == LowPwrRequestSWMask
	s2.state.SFF8636.LowPwrOverride = bin[0x5D]&LowPwrOverrideMask == LowPwrOverrideMask

	// page 0x00
	s2.state.VendorName = util.ParseASCIIToString(bin[0x94:0xA3])
	s2.state.VendorPartNumber = util.ParseASCIIToString(bin[0xA8:0xB7])
	s2.state.VendorPartRevision = util.ParseASCIIToString(bin[0xA8:0xB7])
	s2.state.VendorSerialNumber = util.ParseASCIIToString(bin[0xC4:0xD3])

	return s2.state, nil
}

func (s2 *ManagementStrategy) SetAdministrativeInformation() (*pkg.ModuleState, error) {
	bin, err := s2.state.GetPageBin(0x00, 0)
	if err != nil {
		return nil, err
	}

	bin[0x5D] &^= LowPwrRequestSWMask | SoftwareResetMask // clear
	bin[0x5D] |= LowPwrRequestSWMask & util.YesNoByte(s2.state.LowPwrRequestSW)
	bin[0x5D] |= SoftwareResetMask & util.YesNoByte(s2.state.SoftwareReset)

	err = s2.state.WritePageBin(0x00, 0, bin)
	if err != nil {
		return nil, err
	}

	return s2.state, nil
}
