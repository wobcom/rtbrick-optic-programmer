package pkg

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/rtbrick/ssh"
)

// FinIsarState client r/w interface for FinIsar specific settings
// delegates concrete operations to strategy
type FinIsarState struct {
	// handle internal connection handle
	handle *connection.I2cRWHandle
}

// FlexOptixState client r/w interface for FlexOptix specific settings
// delegates concrete operations to strategy
type FlexOptixState struct {
	// handle internal connection handle
	handle *connection.I2cRWHandle
}

// ModuleState is data exchange interface - delegates to strategy
// client can get / set freely from fields - not linked to pages or concrete implementation
// manufacturer specific fields can be read / set if struct is present, only one manufacturer
// will be present at a time
type ModuleState struct {
	// mgmtProtoConcreteStrategy private pointer to concrete Strategy
	mgmtProtoConcreteStrategy ConcreteManagementStrategy
	// handle pointer to connection handle.
	handle *connection.I2cRWHandle

	// FinIsarSpecific public read only pointer to FinIsar manufacturer specific info
	FinIsarSpecific *FinIsarState
	// FlexOptixSpecific public read only pointer to FlexOptix manufacturer specific info
	FlexOptixSpecific *FlexOptixState

	ManagementProtocol string
	// SFF8024Identifier lower mem public read-only sff8024 id field
	SFF8024Identifier uint8
	// SFF8024Revision lower mem public read-only sff8024 revision id field
	SFF8024Revision uint8

	VendorName         string
	VendorPartNumber   string
	VendorPartRevision string
	VendorSerialNumber string
}

func NewModuleState(
	withDefaultStrategyFactory func(state *ModuleState) ConcreteManagementStrategy,
	withConcreteStrategiesFactories []func(state *ModuleState) ConcreteManagementStrategy,
	withHandle *connection.I2cRWHandle,
) *ModuleState {
	m := &ModuleState{
		handle:             withHandle,
		FinIsarSpecific:    nil,
		FlexOptixSpecific:  nil,
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

func (m ModuleState) GetTunableLaserCtrlStatus() (*ModuleState, error) {
	return m.mgmtProtoConcreteStrategy.GetTunableLaserCtrlStatus()
}

func (m ModuleState) SetTunableLaserCtrlStatus(s *ModuleState) (*ModuleState, error) {
	return m.mgmtProtoConcreteStrategy.SetTunableLaserCtrlStatus(s)
}

// Management is implemented by ModuleState by delegating to strategies which have concrete implementations
type Management interface {
	Set(s *ModuleState) (*ModuleState, error)
	Get() (*ModuleState, error)
	GetAdministrativeInformation() (*ModuleState, error)
	SetAdministrativeInformation(s *ModuleState) (*ModuleState, error)
	GetTunableLaserCtrlStatus() (*ModuleState, error)
	SetTunableLaserCtrlStatus(s *ModuleState) (*ModuleState, error)
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

// ConcreteManagementStrategy should implement both Management and SFF8024Compatible interfaces
type ConcreteManagementStrategy interface {
	Management
	SFF8024Compatible
}

// FlexOptixManagement specific settings for FlexOptix modules. flex tune, nominal wavelength control
type FlexOptixManagement interface {
	GetCustomFlexOptixSettings() (*FlexOptixState, error)
	SetCustomFlexOptixSettings(s *FlexOptixState) (*FlexOptixState, error)
}

// FinIsarManagement specific settings for FinIsar modules. TBD.
type FinIsarManagement interface {
	GetCustomFinIsarSettings() (*FinIsarState, error)
	SetCustomFinIsarSettings(s *FinIsarState) (*FinIsarState, error)
}
