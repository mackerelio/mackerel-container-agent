package nodeinfo

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
