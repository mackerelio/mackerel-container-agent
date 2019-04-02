package ecsv3

import (
	"context"
	"regexp"

	"github.com/mackerelio/mackerel-container-agent/metric"
	"github.com/mackerelio/mackerel-container-agent/platform"
	"github.com/mackerelio/mackerel-container-agent/platform/ecsv3/taskmetadata"
	"github.com/mackerelio/mackerel-container-agent/spec"
)

// APIClient interface gets task metadata and task stats
type APIClient interface {
	TaskMetadataGetter
	// TaskStatsGetter
}

type ecsPlatform struct {
	client    APIClient
	isFargate bool
}

// NewECSPlatform creates a new Platform
func NewECSPlatform(baseURL string, isFargate bool, ignoreContainer *regexp.Regexp) (platform.Platform, error) {
	c, err := taskmetadata.NewClient(baseURL, ignoreContainer)
	if err != nil {
		return nil, err
	}
	return &ecsPlatform{
		client:    c,
		isFargate: isFargate,
	}, nil
}

// GetMetricGenerators gets metric generators
func (p *ecsPlatform) GetMetricGenerators() []metric.Generator {
	return nil
}

// GetSpecGenerators gets spec generator
func (p *ecsPlatform) GetSpecGenerators() []spec.Generator {
	return []spec.Generator{
		newSpecGenerator(p.client, p.isFargate),
	}
}

// GetCustomIdentifier gets custom identifier
func (p *ecsPlatform) GetCustomIdentifier(context.Context) (string, error) {
	return "", nil
}

// StatusRunning reports p status is running
func (p *ecsPlatform) StatusRunning(context.Context) bool {
	return false
}
