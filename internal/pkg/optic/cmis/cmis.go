package cmis

import (
	"fmt"

	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/optic/util"
)

type ManagementStrategy struct {
	state *pkg.ModuleState
}

func New(state *pkg.ModuleState) *ManagementStrategy {
	return &ManagementStrategy{
		state: state,
	}
}

func (s2 *ManagementStrategy) GetPageBin(page byte, bank byte) ([]byte, error) {
	handle := s2.state.GetHandle()

	// we do not have one-shot 2 bytes write available
	// currently this is out of spec
	// > Note: In other words, a host needs to write both the BankSelect and the PageSelect register in a single WRITE
	// > transaction, except for the case when BankIndex value in the BankSelect register does not change. In this case
	// > it is sufficient to WRITE to the PageSelect register.
	// > Note: However, modules may also choose to accept the two registers written in two subsequent WRITE
	// > transactions, to work with non-compliant hosts. Note that rainy day scenarios remain unspecified in this case,
	// > and the PageMapping Validity assertion described below is compromised.

	// do bank select first (will not be cleared if incorrect)
	err := handle.Connection.DoI2CSet(handle.I2cBusId, BankSelectRegisterAddress, bank)
	if err != nil {
		panic(BankSelectErrorString)
	}

	// do page select next (will be cleared if page + bank is incorrect)
	err = handle.Connection.DoI2CSet(handle.I2cBusId, PageSelectRegisterAddress, page)
	if err != nil {
		panic(PageSelectErrorString)
	}

	// decode page
	pageStr, err := handle.Connection.GetI2CDump(handle.I2cBusId)
	if err != nil {
		return []byte{}, err
	}
	dumpBin := pkg.ParseI2CDump(*pageStr)

	// check page + bank select was authorized only and only if target page select = page select
	if dumpBin[PageSelectRegisterAddress] != page {
		panic(fmt.Sprintf(PageOrBankNotAvailableTemplateErrorString, page, bank))
	}

	// return if everything went ok
	return dumpBin, nil
}

func (s2 *ManagementStrategy) WritePageByteBin(page byte, bank byte, offset byte, value byte) error {
	handle := s2.state.GetHandle()

	_, err := s2.GetPageBin(page, bank) // select page by read
	if err != nil {
		return err
	}

	// write is authorized if nothing failed
	err = handle.Connection.DoI2CSet(handle.I2cBusId, (int)(offset), value)
	if err != nil {
		panic(RegisterWriteErrorString)
	}

	return nil
}

func (s2 *ManagementStrategy) AcceptsSFF8024(sff8024Identifier byte, sff8024Revision byte) bool {
	return checkSFF8024(sff8024Identifier, sff8024Revision)
}

func (s2 *ManagementStrategy) Set() (*pkg.ModuleState, error) {
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
	dumpBin, err := s2.state.GetPageBin(0x00, 0x00)
	if err != nil {
		return nil, err
	}

	s2.state.ManagementProtocol = "cmis"

	// lower mem
	s2.state.CMIS.MemoryModelPaged = dumpBin[0x02]&MemoryModelMask == 0 // 0 is paged memory, 1 is flat
	s2.state.CMIS.SteppedConfigOnly = dumpBin[0x02]&SteppedConfigOnlyMask == 0
	s2.state.CMIS.I2CMciMaxSpeedKhz = I2CMciMaxSpeedToKhz[dumpBin[0x02]&I2CMciMaxSpeedMask]
	s2.state.CMIS.SPIMciMaxSpeedKhz = SPIMciMaxSpeedToKhz[dumpBin[0x02]&SPIMCIMaxSpeedMask]
	AutoCommissioning := dumpBin[0x02] & AutoCommissioningMask
	s2.state.CMIS.AutoCommissioningNone = AutoCommissioning == 0b0000
	s2.state.CMIS.AutoCommissioningRegular = AutoCommissioning == 0b0001
	s2.state.CMIS.AutoCommissioningHot = AutoCommissioning == 0b0010 // 0x11 is reserved

	// page 00
	s2.state.VendorName = util.ParseASCIIToString(dumpBin[0x81:0x90])
	s2.state.CMIS.VendorOUI = dumpBin[0x91:0x93]
	s2.state.VendorPartNumber = util.ParseASCIIToString(dumpBin[0x94:0xA3])
	s2.state.VendorPartRevision = util.ParseASCIIToString(dumpBin[0xA4:0xA5])
	s2.state.VendorSerialNumber = util.ParseASCIIToString(dumpBin[0xA6:0xB5])
	s2.state.CMIS.DateCode = util.ParseASCIIToString(dumpBin[0xB6:0xBD])
	s2.state.CMIS.CLEICode = util.ParseASCIIToString(dumpBin[0xBE:0xC7])
	s2.state.CMIS.PowerClass = PowerClassToInt[dumpBin[0xC8]&ModulePowerClassMask]
	s2.state.CMIS.MaxPowerWatts = 0.25 * float64(dumpBin[0xC9]) // byte is interpreted as uint8. unit is ceil of quarter-watts

	s2.state.CMIS.SupportedMediaLanes = make(map[int]bool)
	// fetching media lane support per lane
	var mask byte = 0b1000_0000
	for i := 8; i >= 1; i -= 1 { // MSB countdown
		s2.state.CMIS.SupportedMediaLanes[i] = dumpBin[0xD2]&mask == 0 // supported == 0, unsupported == 1
		mask >>= 1
	}
	s2.state.CMIS.FarEndDetachableMedia = dumpBin[0xD3]&FarEndConfigurationMask == 0b0000_0000
	s2.state.CMIS.FarEnd1LaneModule = dumpBin[0xD3]&FarEndConfigurationMask == 0b0000_0001
	s2.state.CMIS.FarEnd2LanesModule = dumpBin[0xD3]&FarEndConfigurationMask == 0b000_0110
	s2.state.CMIS.FarEnd4LanesModule = dumpBin[0xD3]&FarEndConfigurationMask == 0b000_0011
	s2.state.CMIS.FarEnd8LanesModule = dumpBin[0xD3]&FarEndConfigurationMask == 0b000_0010
	s2.state.CMIS.FarEnd16LanesModule = dumpBin[0xD3]&FarEndConfigurationMask == 0b001_1011
	s2.state.CMIS.MediaInterfaceDescription = MediaInterfaceToStr[dumpBin[0xD4]]

	return s2.state, nil
}

func (s2 *ManagementStrategy) SetAdministrativeInformation() (*pkg.ModuleState, error) {
	// noop
	return s2.state, nil
}
