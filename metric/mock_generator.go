package metric

import (
	"context"

	mackerel "github.com/mackerelio/mackerel-client-go"
)

// MockGenerator represents a mock metric generator
type MockGenerator struct {
	values       Values
	errValues    error
	graphDefs    []*mackerel.GraphDefsParam
	errGraphDefs error
}

// NewMockGenerator creates a new mock metric generator
func NewMockGenerator(values Values, errValues error, graphDefs []*mackerel.GraphDefsParam, errGraphDefs error) Generator {
	return &MockGenerator{values, errValues, graphDefs, errGraphDefs}
}

// Generate generates metric values
func (g *MockGenerator) Generate(context.Context) (Values, error) {
	return g.values, g.errValues
}

// GetGraphDefs gets graph definitions
func (g *MockGenerator) GetGraphDefs(context.Context) ([]*mackerel.GraphDefsParam, error) {
	return g.graphDefs, g.errGraphDefs
}
