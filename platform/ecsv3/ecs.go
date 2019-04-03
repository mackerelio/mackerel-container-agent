package ecsv3

import (
	"context"
	"regexp"
	"strings"

	"github.com/mackerelio/golib/logging"
	"github.com/mackerelio/mackerel-container-agent/metric"
	"github.com/mackerelio/mackerel-container-agent/platform"
	"github.com/mackerelio/mackerel-container-agent/platform/ecsv3/taskmetadata"
	"github.com/mackerelio/mackerel-container-agent/spec"
)

var logger = logging.GetLogger("ecs")

// APIClient interface gets task metadata and task stats
type APIClient interface {
	TaskMetadataGetter
	TaskStatsGetter
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
	return []metric.Generator{
		newMetricGenerator(p.client),
	}
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
func (p *ecsPlatform) StatusRunning(ctx context.Context) bool {
	meta, err := p.client.GetTaskMetadata(ctx)
	if err != nil {
		logger.Warningf("failed to get metadata: %s", err)
		return false
	}
	return isRunning(meta.KnownStatus)
}

func isRunning(status string) bool {
	return strings.EqualFold("running", status)
}
