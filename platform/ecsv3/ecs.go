package ecsv3

import (
	"context"
	"errors"
	"regexp"

	"github.com/mackerelio/mackerel-container-agent/metric"
	"github.com/mackerelio/mackerel-container-agent/platform"
	"github.com/mackerelio/mackerel-container-agent/spec"
)

type ecsPlatform struct{}

// NewECSPlatform creates a new Platform
func NewECSPlatform(ignoreContainer *regexp.Regexp) (platform.Platform, error) {
	return nil, nil
}

// GetMetricGenerators gets metric generators
func (p *ecsPlatform) GetMetricGenerators() []metric.Generator {
	return nil
}

// GetSpecGenerators gets spec generator
func (p *ecsPlatform) GetSpecGenerators() []spec.Generator {
	return []spec.Generator{
		newSpecGenerator(nil, false), // TODO
	}
}

// GetCustomIdentifier gets custom identifier
func (p *ecsPlatform) GetCustomIdentifier(context.Context) (string, error) {
	return "", errors.New("not implemented yet")
}

// StatusRunning reports p status is running
func (p *ecsPlatform) StatusRunning(context.Context) bool {
	return false
}
