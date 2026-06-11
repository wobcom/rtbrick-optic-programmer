package pkg

import (
	"encoding/json"
	"fmt"
	"reflect"

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
}

// CMISOnlyExtension CMIS specific information
type CMISOnlyExtension struct {
	Active bool
}

// SFF8636OnlyExtension SFF8636 specific information
type SFF8636OnlyExtension struct {
	Active bool
}

// ModuleState is data exchange interface - delegates to strategy
// client can get / set freely from fields - not linked to pages or concrete implementation
// manufacturer specific fields can be read / set if struct is present, only one manufacturer
// will be present at a time
type ModuleState struct {
	mgmtProtoConcreteStrategy             ConcreteManagementStrategy // private pointer to concrete Strategy
	mgmtProtoExtensionsConcreteStrategies []ConcreteExtensionManagementStrategy
	handle                                *connection.I2cRWHandle // private pointer to connection handle.

	FinIsarCMISExtension      FinIsarCMISExtension
	FlexOptixSFF8636Extension FlexOptixSFF8636Extension
	SFF8636OnlyExtension      SFF8636OnlyExtension
	CMISOnlyExtension         CMISOnlyExtension

	// lower mem region
	ManagementProtocol     string
	SFF8024Identifier      uint8 // SFF8024Identifier lower mem public read-only sff8024 id field
	SFF8024Revision        uint8 // SFF8024Revision lower mem public read-only sff8024 revision id field
	SoftwareReset          bool
	EnableHighPowerClass8  bool // TODO check if PowerClasses are present in CMIS too
	EnableHighPowerClass57 bool
	LowPwrRequestSW        bool
	LowPwrOverride         bool

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

func (m ModuleState) GetPageBin(page byte) ([]byte, error) {
	pageStr, err := m.handle.Connection.GetI2CDump(m.handle.I2cBusId, page)
	if err != nil {
		return []byte{}, err
	}
	return ParseI2CDump(*pageStr), nil
}

func (m ModuleState) WritePageBin(page byte, offset byte, value byte) error {
	// TODO implement me
	panic("not implemented.")
}

func (m ModuleState) ToJson() ([]byte, error) {
	marshal, err := json.MarshalIndent(m, "", "\t")
	if err != nil {
		return nil, err
	}

	return marshal, nil
}

func (m ModuleState) Set(s *ModuleState) (*ModuleState, error) {
	return m.mgmtProtoConcreteStrategy.Set(s)
}

func (m ModuleState) Get() (*ModuleState, error) {
	return m.mgmtProtoConcreteStrategy.Get()
}

func (m ModuleState) GetAdministrativeInformation() (*ModuleState, error) {
	return m.mgmtProtoConcreteStrategy.GetAdministrativeInformation()
}

func (m ModuleState) SetAdministrativeInformation(s *ModuleState) (*ModuleState, error) {
	return m.mgmtProtoConcreteStrategy.SetAdministrativeInformation(s)
}

func (m ModuleState) SetExtensionsState(s *ModuleState) (*ModuleState, error) {
	for _, e := range m.mgmtProtoExtensionsConcreteStrategies {
		_, err := e.SetExtensionState(s)
		if err != nil {
			return nil, err
		}
	}
	return &m, nil
}

func (m ModuleState) GetExtensionsState() (*ModuleState, error) {
	for _, e := range m.mgmtProtoExtensionsConcreteStrategies {
		_, err := e.GetExtensionState() // force refresh
		if err != nil {
			return nil, err
		}
	}
	return &m, nil
}

// Management is implemented by ModuleState by delegating to strategies which have concrete implementations
type Management interface {
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

// DirectPageAccess Allows direct read/write access to pages
type DirectPageAccess interface {
	GetPageBin(page byte) ([]byte, error)
	WritePageBin(page byte, offset byte, value byte) error
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
