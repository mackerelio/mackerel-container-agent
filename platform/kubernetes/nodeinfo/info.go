package nodeinfo

import (
	"runtime"

	"github.com/mackerelio/go-osstat/memory"
)

// Generator interface gets node information
type Generator interface {
	GetInfo() (memTotal float64, cpuCores float64, err error)
}

// NewGenerator returns node info generator
func NewGenerator() *RealGenerator {
	return &RealGenerator{}
}

// RealGenerator is a real Generator
type RealGenerator struct{}

// GetInfo retrieves information and return it
func (r *RealGenerator) GetInfo() (memTotal float64, cpuCores float64, err error) {
	memory, err := memory.Get()
	if err != nil {
		return 0, 0, err
	}
	return float64(memory.Total), float64(runtime.NumCPU()), nil
}

// MockGenerator is mock for testing
type MockGenerator struct {
	mockCPUCores float64
	mockMemTotal float64
	mockErr      error
}

// GetInfo returns mock response
func (m *MockGenerator) GetInfo() (float64, float64, error) {
	return m.mockMemTotal, m.mockCPUCores, m.mockErr
}

// NewMockGenerator is constructor for mock
func NewMockGenerator(mockMemTotal float64, mockCPUCores float64, mockErr error) *MockGenerator {
	return &MockGenerator{
		mockCPUCores: mockCPUCores,
		mockMemTotal: mockMemTotal,
		mockErr:      mockErr,
	}
}
