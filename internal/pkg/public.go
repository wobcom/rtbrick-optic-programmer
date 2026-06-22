package pkg

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/optic/util"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/rtbrick/ssh"
)

// FinIsarCMISExtension client r/w interface for FinIsar specific settings
// delegates concrete operations to strategy
type FinIsarCMISExtension struct {
	Active bool
}

// FlexOptixSFF8636Extension client r/w interface for FlexOptix specific settings
// delegates concrete operations to strategy
type FlexOptixSFF8636Extension struct {
	Active bool

	// FlexOptix has page 04h and 12h copied from CMIS
	LaserCapabilities      CMISLaserCapabilitiesAdvertising
	TunableLaserCtrlStatus CMISBankedTunableLaserControlAndStatus // n=1 only, up to media 8 lanes support
}

// CMISOnlyExtension CMIS specific information
type CMISOnlyExtension struct {
	Active bool

	// lower mem
	MemoryModelPaged         bool // MemoryModelPaged is true if paged, false if flat (flat is lower + page 0x00 only)
	SteppedConfigOnly        bool // SteppedConfigOnly true if all types of reconfiguration (step by step hot + regular), false if step by step / none autocomm.
	I2CMciMaxSpeedKhz        int  // I2CMciMaxSpeedKhz Maximum I2C MCI interface speed in Khz
	SPIMciMaxSpeedKhz        int  // SPIMciMaxSpeedKhz Maximum MCI SPI interface speed in Khz
	AutoCommissioningNone    bool // AutoCommissioningNone true if no auto-commissioning is supported
	AutoCommissioningRegular bool //  AutoCommissioningRegular true if regular auto-commissioning is supported (Affects ApplyDPInit)
	AutoCommissioningHot     bool // AutoCommissioningHot true if only hot auto-commissioning is supported (Affects ApplyImmediate)

	// page 00, all
	VendorOUI     []byte
	DateCode      string
	CLEICode      string
	PowerClass    int
	MaxPowerWatts float64 // MaxPowerWatts is in multiples of 0.25W, ceil.

	// page 00, copper and active
	// CableAssemblyLengthMeters uint
	// ConnectorType             byte
	// AttenuationAt5Ghz uint8
	// AttenuationAt7Ghz uint8
	// AttenuationAt12p9Ghz uint8
	// AttenuationAt25p8Ghz uint8
	// AttenuationAt53p1Ghz uint8

	// page 00, cont.
	SupportedMediaLanes   map[int]bool
	FarEndDetachableMedia bool
	FarEnd1LaneModule     bool
	FarEnd2LanesModule    bool
	FarEnd4LanesModule    bool
	FarEnd8LanesModule    bool
	FarEnd16LanesModule   bool
	MediaInterface        string

	// page 01, optional
	SupportedControls CMISSupportedControlsAdvertising `json:",omitzero"`

	// supported flags tbd.
	// supported monitors tbd.
	// supported configuration and signal integrity controls adv tbd.
	// CDB messaging support advertisement tbd.
	// additional durations adv tbd.
	// host lane polarity inversion adv tbd.
	// supported pages and banks adv tbd.
	// NAD banks + media lane assignment + additional app adv tbd.
	// misc feature adv tbd.

	// page 12
	LaserCapabilities      CMISLaserCapabilitiesAdvertising          `json:",omitzero"`
	TunableLaserCtrlStatus [4]CMISBankedTunableLaserControlAndStatus `json:",omitzero"` // CMIS 5.3 defines 4 banks max. (0-3 32 lanes)
}

type CMISSupportedControlsAdvertising struct {
	ModuleInactiveFirmwareMajorRevision uint8
	ModuleInactiveFirmwareMinorRevision uint8
	ModuleHardwareMajorRevision         uint8
	ModuleHardwareMinorRevision         uint8
	// link length support tbd.
	NominalWavelengthNm   float64 // NominalWavelengthNm at room temp.
	WavelengthToleranceNm float64 // WavelengthToleranceNm worst case tolerance around nominal
	MaximumBankSupported  byte    // MaximumBankSupported upper bank limit (0, 0-1, 0-3 for 8, 16, 32 lanes respectively)
	// module characteristics adv. tbd.
	WavelengthIsControllable      bool
	TransmitterIsTunable          bool
	SquelchMethodTx               byte
	ForcedSquelchTxSupported      bool
	AutoSquelchDisableTxSupported bool
	AutoSquelchDisableRxSupported bool
	OutputDisableTxSupported      bool
	OutputDisableRxSupported      bool
	InputPolarityFlipTxSupported  bool
	OutputPolarityFlipRxSupported bool
	BankBroadcastSupported        bool
}

type CMISLaserCapabilitiesAdvertising struct {
	// page 04h laser capabilities adv
	// is channel-based grid tuning supported on these frequencies
	SupportedFrequencies map[string]bool

	// is fine-tuning supported in the vicinity of an on-grid channel
	FineTuningSupported bool

	// S16 encoded lowest N for spacing for each freq
	GridLowChannel map[string]int16

	// S16 encoded higher N for spacing for each freq
	GridHighChannel map[string]int16

	// fine-tuning res, 0.001 Ghz increments
	FineTuningResolution uint16
	FineTuningLowOffset  int16
	FineTuningHighOffset int16

	// programmable output power
	ProgOutputPowerPerLaneSupported bool  // per-lane programmable y/n
	ProgOutputPowerMin              int16 // 0.001 dBm increments min power
	ProgOutputPowerMax              int16 // 0.001 dBm increments max power
}

const (
	CMISGridSpacing3p125Ghz     = 0b0000
	CMISGridSpacing6p25Ghz      = 0b0001
	CMISGridSpacing12p5Ghz      = 0b0010
	CMISGridSpacing25Ghz        = 0b0011
	CMISGridSpacing50Ghz        = 0b0100
	CMISGridSpacing100Ghz       = 0b0101
	CMISGridSpacing33Ghz        = 0b0110
	CMISGridSpacing75Ghz        = 0b0111
	CMISGridSpacing150Ghz       = 0b1000
	CMISGridSpacingNotAvailable = 0b1111
)

var CMISGridSpacingToFloatGhzMap = map[byte]float64{
	CMISGridSpacing3p125Ghz: 3.125,
	CMISGridSpacing6p25Ghz:  6.25,
	CMISGridSpacing12p5Ghz:  12.5,
	CMISGridSpacing25Ghz:    25.0,
	CMISGridSpacing50Ghz:    50.0,
	CMISGridSpacing100Ghz:   100.0,
	CMISGridSpacing33Ghz:    33.0,
	CMISGridSpacing75Ghz:    75.0,
	CMISGridSpacing150Ghz:   150.0,
}

var FloatGhzToCMISGridSpacing = map[string]byte{ // cannot use float as key since not serializable
	"3.125":   CMISGridSpacing3p125Ghz,
	"6.250":   CMISGridSpacing6p25Ghz,
	"12.500":  CMISGridSpacing12p5Ghz,
	"25.000":  CMISGridSpacing25Ghz,
	"50.000":  CMISGridSpacing50Ghz,
	"100.000": CMISGridSpacing100Ghz,
	"33.000":  CMISGridSpacing33Ghz,
	"75.000":  CMISGridSpacing75Ghz,
	"150.000": CMISGridSpacing150Ghz,
}

var MultiplierMap = map[string]int{
	"3.125":   3.125e9,
	"6.250":   6.25e9,
	"12.500":  12.5e9,
	"25.000":  25.0e9,
	"50.000":  50.0e9,
	"100.000": 100.0e9,
	"33.000":  33.0e9,
	"75.000":  75.0e9,
	"150.000": 150.0e9,
}

const (
	CMISGridSpacingTxMask            = 0b1111_0000
	CMISFineTuningEnableTxMask       = 0b0000_0001
	CMISTuningInProgressTxMask       = 0b0000_0010
	CMISWavelengthUnlockStatusTxMask = 0b0000_0001

	CMISTargetOutputPowerOORFlagTxMask = 0b0010_0000
	CMISFineTuningOutOfRangeFlagTxMask = 0b0001_0000
	CMISTuningNotAcceptedFlagTxMask    = 0b0000_1000
	CMISInvalidChannelNumberFlagTxMask = 0b0000_0100
	CMISWavelengthUnlockedFlagTxMask   = 0b0000_0010
	CMISTuningCompleteFlagTxMask       = 0b0000_0001
)

type CMISBankedTunableLaserControlAndStatus struct {
	// page 12h tunable laser control and status

	// Grid
	GridSpacingTx      [8]byte    // selected grid spacing of media lanes 1-8 OF BANK
	GridSpacingTxRO    [8]float64 // read only float64 ghz value
	FineTuningEnableTx [8]bool    // for each lane

	// tuning and status
	ChannelNumberTx         [8]int16  // S16 selected N - channel number for media lane 1-8 OF BANK
	FineTuningOffsetTx      [8]int16  // S16 fine-tuning frequency offset for media lane 1-8 OF BANK in offsets of 0.001 Ghz
	CurrentLaserFrequencyTx [8]uint32 // U32 current frequency for media lane 1-8 OF BANK in units of 0.001 Ghz

	// power
	TargetOutputPowerTx [8]int16 // s16 programmable output power for all media lanes IN BANK units of 0.01dBm

	// lock status
	TuningInProgressTx     [8]bool // whether tuning is in progress on all media lanes IN BANK
	WaveLengthUnlockStatus [8]bool // unlocked status indication for laser on all media lanes IN BANK

	// latched flags, cleared on module read
	LaserTuningFlagSummaryTx   [8]bool
	TargetOutputPowerOORFlagTx [8]bool // indicates whether target output power value was entered for media lane
	FineTuningOutOfRangeFlagTx [8]bool // indicates whether fine-tuning target value was outside range
	TuningNotAcceptedFlagTx    [8]bool // indicates a failed tuning operation for media lane
	InvalidChannelNumberFLagTx [8]bool // required channel number not in advertised range of spacing
	WavelengthUnlockedFlagTx   [8]bool
	TuningCompleteFlagTx       [8]bool // tuning has been completed y/n

	// masks for interrupt generation suppression
	TargetOutputPowerOORMaskTx      [8]bool
	FineTuningPowerOutOfRangeMaskTx [8]bool
	TuningNotAcceptedMaskTx         [8]bool
	InvalidChannelMaskTx            [8]bool
	WavelengthUnlockedMaskTx        [8]bool
	TuningCompleteMaskTx            [8]bool
}

// SFF8636OnlyExtension SFF8636 specific information
type SFF8636OnlyExtension struct {
	Active bool

	// lower mem
	EnableHighPowerClass8  bool
	EnableHighPowerClass57 bool
	LowPwrRequestSW        bool
	LowPwrOverride         bool
}

// ModuleState is data exchange interface - delegates to strategy
// client can get / set freely from fields - not linked to pages or concrete implementation
// manufacturer specific fields can be read / set if struct is present, only one manufacturer
// will be present at a time
type ModuleState struct {
	mgmtProtoConcreteStrategy             ConcreteManagementStrategy // private pointer to concrete Strategy
	mgmtProtoExtensionsConcreteStrategies []ConcreteExtensionManagementStrategy
	handle                                *connection.I2cRWHandle // private pointer to connection handle.

	FinIsarCMISExtension      FinIsarCMISExtension      `json:",omitzero"`
	FlexOptixSFF8636Extension FlexOptixSFF8636Extension `json:",omitzero"`
	SFF8636OnlyExtension      SFF8636OnlyExtension      `json:",omitzero"`
	CMISOnlyExtension         CMISOnlyExtension         `json:",omitzero"`

	// lower mem region
	ManagementProtocol string
	SFF8024Identifier  uint8 // SFF8024Identifier lower mem public read-only sff8024 id field
	SFF8024Revision    uint8 // SFF8024Revision lower mem public read-only sff8024 revision id field
	SoftwareReset      bool

	// page 00 region, common info between
	// CMIS and SFF8636 but addresses differ
	VendorName         string
	VendorPartNumber   string
	VendorPartRevision string
	VendorSerialNumber string
}

func NewModuleState(
	withDefaultStrategyFactory func(state *ModuleState) ConcreteManagementStrategy,
	withConcreteStrategiesFactories []func(state *ModuleState) ConcreteManagementStrategy,
	withConcreteExtensionStrategiesFactories []func(state *ModuleState) ConcreteExtensionManagementStrategy,
	withHandle *connection.I2cRWHandle) *ModuleState {
	m := &ModuleState{
		handle:             withHandle,
		ManagementProtocol: "unknown",
	}
	m.mgmtProtoConcreteStrategy = withDefaultStrategyFactory(m) // self-referential

	m, err := m.GetAdministrativeInformation()
	if err != nil {
		panic("Failed to query module for basic administrative information")
	}

	for _, s := range withConcreteStrategiesFactories {
		if s(m).AcceptsSFF8024(m.SFF8024Identifier, m.SFF8024Revision) {
			m.mgmtProtoConcreteStrategy = s(m)
		}
	}

	if reflect.TypeOf(m.mgmtProtoConcreteStrategy) == reflect.TypeOf(withDefaultStrategyFactory(m)) {
		panic(fmt.Sprintf("Unknown SFF8024 Management interface type %x", m.SFF8024Identifier))
	}

	// re-query for more advanced admin info
	// once we have the concrete management strategy confirmed
	m, err = m.GetAdministrativeInformation()
	if err != nil {
		panic("Failed to query module for basic administrative information")
	}

	for _, s := range withConcreteExtensionStrategiesFactories {
		extensionStrategy := s(m)
		if extensionStrategy.SFF8024IsCompatibleWithProtocolExtension(m.SFF8024Identifier, m.SFF8024Revision) &&
			extensionStrategy.ManufacturerIsCompatibleWithProtocolExtension(m.VendorName) {
			_, err := extensionStrategy.Activate()
			if err != nil {
				panic("failed to activate protocol extension")
			}
			m.mgmtProtoExtensionsConcreteStrategies = append(
				m.mgmtProtoExtensionsConcreteStrategies,
				extensionStrategy,
			)
		}
	}

	return m
}

func (m *ModuleState) GetHandle() *connection.I2cRWHandle {
	return m.handle
}

func (m *ModuleState) GetPageBin(page byte, bank byte) ([]byte, error) {
	return m.mgmtProtoConcreteStrategy.GetPageBin(page, bank)
}

func (m *ModuleState) WritePageByteBin(page byte, bank byte, offset byte, value byte) error {
	return m.mgmtProtoConcreteStrategy.WritePageByteBin(page, bank, offset, value)
}

func (m *ModuleState) WritePageBin(page byte, bank byte, new []byte) error {
	old, err := m.GetPageBin(page, bank)
	if err != nil {
		return err
	}

	for offset, newValue := range util.BinDiffIterator(old, new) {
		err := m.WritePageByteBin(page, bank, offset, newValue)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *ModuleState) ToJson() ([]byte, error) {
	marshal, err := json.MarshalIndent(m, "", "\t")
	if err != nil {
		return nil, err
	}

	return marshal, nil
}

func (m *ModuleState) Set(s *ModuleState) (*ModuleState, error) {
	return m.mgmtProtoConcreteStrategy.Set(s)
}

func (m *ModuleState) Get() (*ModuleState, error) {
	_, err := m.mgmtProtoConcreteStrategy.Get()
	if err != nil {
		return nil, err
	}
	_, err = m.GetExtensionsState()
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (m *ModuleState) GetAdministrativeInformation() (*ModuleState, error) {
	return m.mgmtProtoConcreteStrategy.GetAdministrativeInformation()
}

func (m *ModuleState) SetAdministrativeInformation(s *ModuleState) (*ModuleState, error) {
	return m.mgmtProtoConcreteStrategy.SetAdministrativeInformation(s)
}

func (m *ModuleState) SetExtensionsState(s *ModuleState) (*ModuleState, error) {
	for _, e := range m.mgmtProtoExtensionsConcreteStrategies {
		_, err := e.SetExtensionState(s)
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

func (m *ModuleState) GetExtensionsState() (*ModuleState, error) {
	for _, e := range m.mgmtProtoExtensionsConcreteStrategies { // multiple extensions may be active
		_, err := e.GetExtensionState() // force refresh
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

// Management is implemented by ModuleState by delegating to strategies which have concrete implementations
type Management interface {
	GetPageBin(page byte, bank byte) ([]byte, error)                      // GetPageBin page is mandatory, bank is optional as it is only used in CMIS
	WritePageByteBin(page byte, bank byte, offset byte, value byte) error // WritePageByteBin page is mandatory, bank is optional as it is only used in CMIS
	Set(s *ModuleState) (*ModuleState, error)
	Get() (*ModuleState, error)
	GetAdministrativeInformation() (*ModuleState, error)
	SetAdministrativeInformation(s *ModuleState) (*ModuleState, error)
}

// ProtocolExtensionManagement is a generic interface for protocol extensions,
// be it protocol specific or manufacturer specific. I've avoided using Generics for
// the state struct and instead opted to use for an all-containing struct since there's
// no need for the added complexity of Generics in this case.
type ProtocolExtensionManagement interface {
	GetExtensionState() (*ModuleState, error)
	SetExtensionState(s *ModuleState) (*ModuleState, error)
	Activate() (*ModuleState, error)
}

// SFF8024Compatible to be implemented by concrete strategies
type SFF8024Compatible interface {
	// AcceptsSFF8024 tells if the strategy is compatible with the sff8024 type
	AcceptsSFF8024(sff8024Identifier byte, sff8024Revision byte) bool
}

type ManufacturerIsCompatibleWithProtocolExtension interface {
	ManufacturerIsCompatibleWithProtocolExtension(manufacturer string) bool
}

type SFF8024IsCompatibleWithProtocolExtension interface {
	SFF8024IsCompatibleWithProtocolExtension(sff8024Identifier byte, sff8024Revision byte) bool
}

// ConcreteManagementStrategy should implement both Management and SFF8024Compatible interfaces
type ConcreteManagementStrategy interface {
	Management
	SFF8024Compatible
}

type ConcreteExtensionManagementStrategy interface {
	ProtocolExtensionManagement
	ManufacturerIsCompatibleWithProtocolExtension
	SFF8024IsCompatibleWithProtocolExtension
}
