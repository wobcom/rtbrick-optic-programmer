package pkg

import (
	"encoding/json"
	"fmt"

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
	// Handle pointer to connection Handle.
	Handle *connection.I2cRWHandle

	// FinIsarSpecific public read only pointer to FinIsar manufacturer specific info
	FinIsarSpecific *FinIsarState
	// FlexOptixSpecific public read only pointer to FlexOptix manufacturer specific info
	FlexOptixSpecific *FlexOptixState

	// SFF8024Identifier public read-only sff8024 id field
	SFF8024Identifier byte
}

func NewModuleState(
	withDefaultStrategy ConcreteManagementStrategy,
	withConcreteStrategies []ConcreteManagementStrategy,
	withHandle *connection.I2cRWHandle,
) *ModuleState {
	m := &ModuleState{
		Handle:                    withHandle,
		mgmtProtoConcreteStrategy: withDefaultStrategy,
		FinIsarSpecific:           nil,
		FlexOptixSpecific:         nil,
	}

	m, err := m.GetAdministrativeInformation()
	if err != nil {
		panic("Failed to query module for basic administrative information")
	}

	for _, s := range withConcreteStrategies {
		if s.Accepts(m.SFF8024Identifier) {
			m.mgmtProtoConcreteStrategy = s
		}
	}

	if m.mgmtProtoConcreteStrategy == withDefaultStrategy {
		panic(fmt.Sprintf("Unknown SFF8024 Management interface type %x", m.SFF8024Identifier))
	}

	return m
}

func (m *ModuleState) SetStrategy(strategy ConcreteManagementStrategy) {
	m.mgmtProtoConcreteStrategy = strategy
}

func Json(m *ModuleState) {
	ok, err := m.mgmtProtoConcreteStrategy.Get()
	if err != nil {
		return
	}
	marshal, err := json.Marshal(ok)
	if err != nil {
		return
	}
	fmt.Println(string(marshal))
}

func (m *ModuleState) Set(s *ModuleState) (*ModuleState, error) {
	return m.mgmtProtoConcreteStrategy.Set(s)
}

func (m *ModuleState) Get() (*ModuleState, error) {
	return m.mgmtProtoConcreteStrategy.Get()
}

func (m *ModuleState) GetAdministrativeInformation() (*ModuleState, error) {
	return m.mgmtProtoConcreteStrategy.GetAdministrativeInformation()
}

func (m *ModuleState) SetAdministrativeInformation(s *ModuleState) (*ModuleState, error) {
	return m.mgmtProtoConcreteStrategy.SetAdministrativeInformation(s)
}

func (m *ModuleState) GetTunableLaserCtrlStatus() (*ModuleState, error) {
	return m.mgmtProtoConcreteStrategy.GetTunableLaserCtrlStatus()
}

func (m *ModuleState) SetTunableLaserCtrlStatus(s *ModuleState) (*ModuleState, error) {
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

// SFF8024Compatible to be implemented by concrete strategies
type SFF8024Compatible interface {
	// Accepts tells if the strategy is compatible with the sff8024 type
	Accepts(sff8024Identifier byte) bool
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
