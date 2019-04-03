package ecsv3

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/mackerelio/golib/logging"
	"github.com/mackerelio/mackerel-container-agent/metric"
	"github.com/mackerelio/mackerel-container-agent/platform"
	"github.com/mackerelio/mackerel-container-agent/platform/ecsv3/taskmetadata"
	"github.com/mackerelio/mackerel-container-agent/provider"
	"github.com/mackerelio/mackerel-container-agent/spec"
)

const (
	executionEnvFargate = "AWS_ECS_FARGATE"
	executionEnvEC2     = "AWS_ECS_EC2"
)

var logger = logging.GetLogger("ecs")

// APIClient interface gets task metadata and task stats
type APIClient interface {
	TaskMetadataGetter
	TaskStatsGetter
}

type ecsPlatform struct {
	client   APIClient
	provider provider.Type
}

// NewECSPlatform creates a new Platform
func NewECSPlatform(metadataURI string, executionEnv string, ignoreContainer *regexp.Regexp) (platform.Platform, error) {
	c, err := taskmetadata.NewClient(metadataURI, ignoreContainer)
	if err != nil {
		return nil, err
	}

	p, err := resolveProvider(executionEnv)
	if err != nil {
		return nil, err
	}

	return &ecsPlatform{
		client:   c,
		provider: p,
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
		newSpecGenerator(p.client, p.provider),
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

func resolveProvider(executionEnv string) (provider.Type, error) {
	switch executionEnv {
	case executionEnvFargate:
		return provider.Fargate, nil
	case executionEnvEC2:
		return provider.ECS, nil
	default:
		return provider.Type("UNKNOWN"), errors.New("unknown exectution env")
	}
}
