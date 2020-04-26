package none

import (
	"context"

	"github.com/mackerelio/mackerel-container-agent/metric"
	"github.com/mackerelio/mackerel-container-agent/platform"
	"github.com/mackerelio/mackerel-container-agent/spec"
)

type nonePlatform struct{}

// NewNonePlatform creates a new Platform
func NewNonePlatform() (platform.Platform, error) {
	return &nonePlatform{}, nil
}

func (p *nonePlatform) GetMetricGenerators() []metric.Generator {
	return []metric.Generator{}
}

func (p *nonePlatform) GetSpecGenerators() []spec.Generator {
	return []spec.Generator{}
}

func (p *nonePlatform) GetCustomIdentifier(context.Context) (string, error) {
	return "", nil
}

func (p *nonePlatform) StatusRunning(context.Context) bool {
	return true
}
