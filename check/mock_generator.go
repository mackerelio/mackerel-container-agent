package check

import (
	"context"

	mackerel "github.com/mackerelio/mackerel-client-go"
)

// MockGenerator represents a mock check generator
type MockGenerator struct {
	name    string
	memo    string
	results []*Result
	err     error
}

// NewMockGenerator creates a new mock check generator
func NewMockGenerator(name string, memo string, results []*Result, err error) Generator {
	return &MockGenerator{name: name, memo: memo, results: results, err: err}
}

// Config gets check generator config
func (g *MockGenerator) Config() mackerel.CheckConfig {
	return mackerel.CheckConfig{Name: g.name, Memo: g.memo}
}

// Generate generates check report
func (g *MockGenerator) Generate(context.Context) (*Result, error) {
	if len(g.results) == 0 {
		return nil, g.err
	}
	r := g.results[0]
	g.results = g.results[1:]
	return r, g.err
}
