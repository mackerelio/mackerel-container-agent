package spec

import "context"

// MockGenerator represents a mock spec generator
type MockGenerator struct {
	value    interface{}
	errValue error
}

// NewMockGenerator creates a new mock spec generator
func NewMockGenerator(value interface{}, errValue error) Generator {
	return &MockGenerator{value, errValue}
}

// Generate generates spec values
func (g *MockGenerator) Generate(context.Context) (interface{}, error) {
	return g.value, g.errValue
}
