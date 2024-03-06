package spec

import "context"

// MockGenerator represents a mock spec generator
type MockGenerator struct {
	value    any
	errValue error
}

// NewMockGenerator creates a new mock spec generator
func NewMockGenerator(value any, errValue error) Generator {
	return &MockGenerator{value, errValue}
}

// Generate generates spec values
func (g *MockGenerator) Generate(context.Context) (any, error) {
	return g.value, g.errValue
}
