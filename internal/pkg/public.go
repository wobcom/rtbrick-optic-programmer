package pkg

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/optic/util"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/rtbrick/ssh"
)

const (
	Grid75string    = "75.000"
	Grid33String    = "33.000"
	Grid100String   = "100.000"
	Grid50String    = "50.000"
	Grid25String    = "25.000"
	Grid12p5String  = "12.500"
	Grid6p250String = "6.250"
	Grid3p125String = "3.125"
	Grid150String   = "150.000"
)

const (
	CMISGridSpacing3p125Ghz = 0b0000_0000
	CMISGridSpacing6p25Ghz  = 0b0001_0000
	CMISGridSpacing12p5Ghz  = 0b0010_0000
	CMISGridSpacing25Ghz    = 0b0011_0000
	CMISGridSpacing50Ghz    = 0b0100_0000
	CMISGridSpacing100Ghz   = 0b0101_0000
	CMISGridSpacing33Ghz    = 0b0110_0000
	CMISGridSpacing75Ghz    = 0b0111_0000
	CMISGridSpacing150Ghz   = 0b1000_0000
	//	CMISGridSpacingNotAvailable = 0b1111_0000
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
	Grid3p125String: CMISGridSpacing3p125Ghz,
	Grid6p250String: CMISGridSpacing6p25Ghz,
	Grid12p5String:  CMISGridSpacing12p5Ghz,
	Grid25String:    CMISGridSpacing25Ghz,
	Grid50String:    CMISGridSpacing50Ghz,
	Grid100String:   CMISGridSpacing100Ghz,
	Grid33String:    CMISGridSpacing33Ghz,
	Grid75string:    CMISGridSpacing75Ghz,
	Grid150String:   CMISGridSpacing150Ghz,
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

// FinIsarCMISExtension client r/w interface for FinIsar specific settings
// delegates concrete operations to strategy
type FinIsarCMISExtension struct {
	Active bool `json:"active"`
}

// CommonTunableLaserFields is the common interface for tunable lasers,
type CommonTunableLaserFields struct {
	Capabilities CMISLaserCapabilitiesAdvertising         `json:"capabilities,omitzero"`
	CtrlStatus   []CMISBankedTunableLaserControlAndStatus `json:"control_status,omitzero"`
}

// FlexOptixSFF8636Extension client r/w interface for FlexOptix specific settings
// delegates concrete operations to strategy
type FlexOptixSFF8636Extension struct {
	Active bool `json:"active"`

	// FlexOptix has page 04h and 12h copied from CMIS
	TunableLaser CommonTunableLaserFields `json:"tunable_laser"`
}

// CMISOnlyExtension CMIS specific information
type CMISOnlyExtension struct {
	Active bool `json:"active"`

	// lower mem
	MemoryModelPaged         bool `json:"memory_model_paged"`         // MemoryModelPaged is true if paged, false if flat (flat is lower + page 0x00 only)
	SteppedConfigOnly        bool `json:"stepped_config_only"`        // SteppedConfigOnly true if all types of reconfiguration (step by step hot + regular), false if step by step / none autocomm.
	I2CMciMaxSpeedKhz        int  `json:"i2c_mci_max_speed_khz"`      // I2CMciMaxSpeedKhz Maximum I2C MCI interface speed in Khz
	SPIMciMaxSpeedKhz        int  `json:"spi_mci_max_speed_khz"`      // SPIMciMaxSpeedKhz Maximum MCI SPI interface speed in Khz
	AutoCommissioningNone    bool `json:"auto_commissioning_none"`    // AutoCommissioningNone true if no auto-commissioning is supported
	AutoCommissioningRegular bool `json:"auto_commissioning_regular"` //  AutoCommissioningRegular true if regular auto-commissioning is supported (Affects ApplyDPInit)
	AutoCommissioningHot     bool `json:"auto_commissioning_hot"`     // AutoCommissioningHot true if only hot auto-commissioning is supported (Affects ApplyImmediate)

	// page 00, all
	VendorOUI     []byte  `json:"vendor_oui"`
	DateCode      string  `json:"date_code"`
	CLEICode      string  `json:"clei_code"`
	PowerClass    int     `json:"power_class"`
	MaxPowerWatts float64 `json:"max_power_watts"` // MaxPowerWatts is in multiples of 0.25W, ceil.

	// page 00, copper and active
	// CableAssemblyLengthMeters uint
	// ConnectorType             byte
	// AttenuationAt5Ghz uint8
	// AttenuationAt7Ghz uint8
	// AttenuationAt12p9Ghz uint8
	// AttenuationAt25p8Ghz uint8
	// AttenuationAt53p1Ghz uint8

	// page 00, cont.
	SupportedMediaLanes       map[int]bool `json:"supported_media_lanes"`
	FarEndDetachableMedia     bool         `json:"far_end_detachable_media"`
	FarEnd1LaneModule         bool         `json:"far_end_1_lane_module"`
	FarEnd2LanesModule        bool         `json:"far_end_2_lanes_module"`
	FarEnd4LanesModule        bool         `json:"far_end_4_lanes_module"`
	FarEnd8LanesModule        bool         `json:"far_end_8_lanes_module"`
	FarEnd16LanesModule       bool         `json:"far_end_16_lanes_module"`
	MediaInterfaceDescription string       `json:"media_interface_description"`

	// page 01, optional
	SupportedControls CMISSupportedControlsAdvertising `json:"supported_controls,omitzero"`

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
	TunableLaser CommonTunableLaserFields `json:"tunable_laser"`
}

type CMISSupportedControlsAdvertising struct {
	ModuleInactiveFirmwareMajorRevision uint8 `json:"module_inactive_firmware_major_revision"`
	ModuleInactiveFirmwareMinorRevision uint8 `json:"module_inactive_firmware_minor_revision"`
	ModuleHardwareMajorRevision         uint8 `json:"module_hardware_major_revision"`
	ModuleHardwareMinorRevision         uint8 `json:"module_hardware_minor_revision"`
	// link length support tbd.
	NominalWavelengthNm   float64 `json:"nominal_wavelength_nm"`   // NominalWavelengthNm at room temp.
	WavelengthToleranceNm float64 `json:"wavelength_tolerance_nm"` // WavelengthToleranceNm worst case tolerance around nominal
	MaximumBankSupported  byte    `json:"maximum_bank_supported"`  // MaximumBankSupported upper bank limit (0, 0-1, 0-3 for 8, 16, 32 lanes respectively)
	// module characteristics adv. tbd.
	WavelengthIsControllable      bool `json:"wavelength_is_controllable"`
	TransmitterIsTunable          bool `json:"transmitter_is_tunable"`
	SquelchMethodTx               byte `json:"squelch_method_tx"`
	ForcedSquelchTxSupported      bool `json:"forced_squelch_tx_supported"`
	AutoSquelchDisableTxSupported bool `json:"auto_squelch_disable_tx_supported"`
	AutoSquelchDisableRxSupported bool `json:"auto_squelch_disable_rx_supported"`
	OutputDisableTxSupported      bool `json:"output_disable_tx_supported"`
	OutputDisableRxSupported      bool `json:"output_disable_rx_supported"`
	InputPolarityFlipTxSupported  bool `json:"input_polarity_flip_tx_supported"`
	OutputPolarityFlipRxSupported bool `json:"output_polarity_flip_rx_supported"`
	BankBroadcastSupported        bool `json:"bank_broadcast_supported"`
}

type CMISLaserCapabilitiesAdvertising struct {
	// page 04h laser capabilities adv
	// is channel-based grid tuning supported on these frequencies
	SupportedGridSpacings map[string]bool `json:"supported_grid_spacings"`

	// is fine-tuning supported in the vicinity of an on-grid channel
	FineTuningSupported bool `json:"fine_tuning_supported"`

	// S16 encoded lowest N for spacing for each freq
	GridLowChannel map[string]int16 `json:"grid_low_channel"`

	// S16 encoded higher N for spacing for each freq
	GridHighChannel map[string]int16 `json:"grid_high_channel"`

	// fine-tuning res, 0.001 Ghz increments
	FineTuningResolution uint16 `json:"fine_tuning_resolution"`
	FineTuningLowOffset  int16  `json:"fine_tuning_low_offset"`
	FineTuningHighOffset int16  `json:"fine_tuning_high_offset"`

	// programmable output power
	ProgOutputPowerPerLaneSupported bool  `json:"prog_output_power_per_lane_supported"` // per-lane programmable y/n
	ProgOutputPowerMin              int16 `json:"prog_output_power_min"`                // 0.001 dBm increments min power
	ProgOutputPowerMax              int16 `json:"prog_output_power_max"`                // 0.001 dBm increments max power
}

type CMISBankedTunableLaserControlAndStatus struct {
	// page 12h tunable laser control and status

	// Grid
	GridSpacingTx      [8]byte    `json:"-"`                     // selected grid spacing of media lanes 1-8 OF BANK
	GridSpacingTxROGhz [8]float64 `json:"grid_spacing_tx"`       // read only float64 ghz value
	FineTuningEnableTx [8]bool    `json:"fine_tuning_enable_tx"` // for each lane

	// tuning and status
	ChannelNumberTx            [8]int16  `json:"channel_number_tx"`              // S16 selected N - channel number for media lane 1-8 OF BANK
	FineTuningOffsetMhzTx      [8]int16  `json:"fine_tuning_offset_mhz_tx"`      // S16 fine-tuning frequency offset for media lane 1-8 OF BANK in offsets of 0.001 Ghz
	CurrentLaserFrequencyMhzTx [8]uint32 `json:"current_laser_frequency_mhz_tx"` // U32 current frequency for media lane 1-8 OF BANK in units of 0.001 Ghz

	// power
	TargetOutputPowerTx [8]int16 `json:"target_output_power_tx"` // s16 programmable output power for all media lanes IN BANK units of 0.01dBm

	// lock status
	TuningInProgressTx     [8]bool `json:"tuning_in_progress_tx"`     // whether tuning is in progress on all media lanes IN BANK
	WaveLengthUnlockStatus [8]bool `json:"wave_length_unlock_status"` // unlocked status indication for laser on all media lanes IN BANK

	// latched flags, cleared on module read
	LaserTuningFlagSummaryTx   [8]bool `json:"laser_tuning_flag_summary_tx"`
	TargetOutputPowerOORFlagTx [8]bool `json:"target_output_power_oor_flag_tx"`  // indicates whether target output power value was entered for media lane
	FineTuningOutOfRangeFlagTx [8]bool `json:"fine_tuning_out_of_range_flag_tx"` // indicates whether fine-tuning target value was outside range
	TuningNotAcceptedFlagTx    [8]bool `json:"tuning_not_accepted_flag_tx"`      // indicates a failed tuning operation for media lane
	InvalidChannelNumberFlagTx [8]bool `json:"invalid_channel_number_flag_tx"`   // required channel number not in advertised range of spacing
	WavelengthUnlockedFlagTx   [8]bool `json:"wavelength_unlocked_flag_tx"`
	TuningCompleteFlagTx       [8]bool `json:"tuning_complete_flag_tx"` // tuning has been completed y/n

	// masks for interrupt generation suppression
	TargetOutputPowerOORMaskTx      [8]bool `json:"target_output_power_oor_mask_tx"`
	FineTuningPowerOutOfRangeMaskTx [8]bool `json:"fine_tuning_power_out_of_range_mask_tx"`
	TuningNotAcceptedMaskTx         [8]bool `json:"tuning_not_accepted_mask_tx"`
	InvalidChannelMaskTx            [8]bool `json:"invalid_channel_mask_tx"`
	WavelengthUnlockedMaskTx        [8]bool `json:"wavelength_unlocked_mask_tx"`
	TuningCompleteMaskTx            [8]bool `json:"tuning_complete_mask_tx"`
}

// SFF8636OnlyExtension SFF8636 specific information
type SFF8636OnlyExtension struct {
	Active bool `json:"active"`

	// lower mem
	EnableHighPowerClass8  bool `json:"enable_high_power_class_8"`
	EnableHighPowerClass57 bool `json:"enable_high_power_class_57"`
	LowPwrOverride         bool `json:"low_pwr_override"`
}

// ModuleState is data exchange interface - delegates to strategy
// client can get / set freely from fields - not linked to pages or concrete implementation
// manufacturer specific fields can be read / set if struct is present, only one manufacturer
// will be present at a time
type ModuleState struct {
	mgmtProtoConcreteStrategy             ConcreteManagementStrategy // private pointer to concrete Strategy
	mgmtProtoExtensionsConcreteStrategies []ConcreteExtensionManagementStrategy
	handle                                *connection.I2cRWHandle // private pointer to connection handle.

	FinIsarCMIS      FinIsarCMISExtension      `json:"finisar_cmis,omitzero"`
	FlexOptixSFF8636 FlexOptixSFF8636Extension `json:"flexoptix_sff8636,omitzero"`
	SFF8636          SFF8636OnlyExtension      `json:"sff8636,omitzero"`
	CMIS             CMISOnlyExtension         `json:"cmis,omitzero"`

	// lower mem region
	ManagementProtocol string `json:"management_protocol"`
	SFF8024Identifier  uint8  `json:"sff_8024_identifier"` // SFF8024Identifier lower mem public read-only sff8024 id field
	SFF8024Revision    uint8  `json:"sff_8024_revision"`   // SFF8024Revision lower mem public read-only sff8024 revision id field
	LowPwrRequestSW    bool   `json:"low_pwr_request_sw"`
	SoftwareReset      bool   `json:"software_reset"`

	// page 00 region, common info between
	// CMIS and SFF8636 but addresses differ
	VendorName         string `json:"vendor_name"`
	VendorPartNumber   string `json:"vendor_part_number"`
	VendorPartRevision string `json:"vendor_part_revision"`
	VendorSerialNumber string `json:"vendor_serial_number"`
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
		err := m.WritePageByteBin(page, bank, byte(offset), newValue)
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

func (m *ModuleState) Set() (*ModuleState, error) {
	return m.mgmtProtoConcreteStrategy.Set()
}

func (m *ModuleState) Get() (*ModuleState, error) {
	_, err := m.mgmtProtoConcreteStrategy.Get()
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (m *ModuleState) GetAdministrativeInformation() (*ModuleState, error) {
	return m.mgmtProtoConcreteStrategy.GetAdministrativeInformation()
}

func (m *ModuleState) SetAdministrativeInformation() (*ModuleState, error) {
	return m.mgmtProtoConcreteStrategy.SetAdministrativeInformation()
}

func (m *ModuleState) SetExtensionsState() (*ModuleState, error) {
	for _, e := range m.mgmtProtoExtensionsConcreteStrategies {
		_, err := e.SetExtensionState()
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
	Set() (*ModuleState, error)
	Get() (*ModuleState, error)
	GetAdministrativeInformation() (*ModuleState, error)
	SetAdministrativeInformation() (*ModuleState, error)
}

// ProtocolExtensionManagement is a generic interface for protocol extensions,
// be it protocol specific or manufacturer specific. I've avoided using Generics for
// the state struct and instead opted to use for an all-containing struct since there's
// no need for the added complexity of Generics in this case.
type ProtocolExtensionManagement interface {
	GetExtensionState() (*ModuleState, error)
	SetExtensionState() (*ModuleState, error)
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
