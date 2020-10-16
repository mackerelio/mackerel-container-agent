package nodeinfo

import (
	"runtime"

	"github.com/mackerelio/go-osstat/memory"
)

// Generator interface gets node information
type Generator interface {
	Generate() (memTotal float64, cpuCores float64, err error)
}

// NewGenerator returns node info generator
func NewGenerator() Generator {
	return &generator{}
}

// generator is a real Generator
type generator struct{}

// Generate retrieves information and return it
func (r *generator) Generate() (memTotal float64, cpuCores float64, err error) {
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

// Generate returns mock response
func (m *MockGenerator) Generate() (float64, float64, error) {
	return m.mockMemTotal, m.mockCPUCores, m.mockErr
}

// NewMockGenerator is constructor for mock
func NewMockGenerator(mockMemTotal float64, mockCPUCores float64, mockErr error) Generator {
	return &MockGenerator{
		mockCPUCores: mockCPUCores,
		mockMemTotal: mockMemTotal,
		mockErr:      mockErr,
	}
}
