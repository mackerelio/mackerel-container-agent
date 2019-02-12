package platform

import (
	"context"

	"github.com/mackerelio/mackerel-container-agent/metric"
	"github.com/mackerelio/mackerel-container-agent/spec"
)

// Platform interface gets metric values and metadata
type Platform interface {
	GetMetricGenerators() []metric.Generator
	GetSpecGenerators() []spec.Generator
	GetCustomIdentifier(context.Context) (string, error)
	StatusRunning(context.Context) bool
}
