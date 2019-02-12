package ecsawsvpc

import (
	"context"
	"regexp"
	"strings"

	"github.com/mackerelio/golib/logging"

	"github.com/mackerelio/mackerel-container-agent/metric"
	"github.com/mackerelio/mackerel-container-agent/platform"
	"github.com/mackerelio/mackerel-container-agent/platform/ecsawsvpc/taskmetadata"
	"github.com/mackerelio/mackerel-container-agent/spec"
)

var logger = logging.GetLogger("ecsawsvpc")

type ecsAwsvpcPlatform struct {
	client    taskmetadata.Client
	isFargate bool
}

// NewECSAwsvpcPlatform creates a new Platform
func NewECSAwsvpcPlatform(isFargate bool, ignoreContainer *regexp.Regexp) (platform.Platform, error) {
	c, err := taskmetadata.NewClient(taskmetadata.DefaultURL, ignoreContainer)
	if err != nil {
		return nil, err
	}
	return &ecsAwsvpcPlatform{
		client:    c,
		isFargate: isFargate,
	}, nil
}

// GetMetricGenerators gets metric generators
func (p *ecsAwsvpcPlatform) GetMetricGenerators() []metric.Generator {
	return []metric.Generator{
		newMetricGenerator(p.client),
		metric.NewInterfaceGenerator(),
	}
}

// GetSpecGenerators gets spec generator
func (p *ecsAwsvpcPlatform) GetSpecGenerators() []spec.Generator {
	return []spec.Generator{
		newSpecGenerator(p.client, p.isFargate),
	}
}

// GetCustomIdentifier gets custom identifier
func (p *ecsAwsvpcPlatform) GetCustomIdentifier(context.Context) (string, error) {
	return "", nil
}

// StatusRunning reports p status is running
func (p *ecsAwsvpcPlatform) StatusRunning(ctx context.Context) bool {
	meta, err := p.client.GetMetadata(ctx)
	if err != nil {
		logger.Warningf("failed to get metadata: %s", err)
		return false
	}
	return strings.EqualFold("running", meta.KnownStatus)
}
