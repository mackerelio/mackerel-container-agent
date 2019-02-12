package metric

import (
	"context"

	mackerel "github.com/mackerelio/mackerel-client-go"
)

// Values represents metric values
type Values map[string]float64

// Generator interface generates metrics
type Generator interface {
	Generate(context.Context) (Values, error)
	GetGraphDefs(context.Context) ([]*mackerel.GraphDefsParam, error)
}
