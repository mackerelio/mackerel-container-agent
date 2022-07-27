package ecs

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	ecsTypes "github.com/mackerelio/mackerel-container-agent/internal/amazon-ecs-agent/agent/handlers/v2"

	"github.com/mackerelio/golib/logging"

	"github.com/mackerelio/mackerel-container-agent/metric"
	"github.com/mackerelio/mackerel-container-agent/metric/hostinfo"
	"github.com/mackerelio/mackerel-container-agent/platform"
	"github.com/mackerelio/mackerel-container-agent/platform/ecs/taskmetadata"
	"github.com/mackerelio/mackerel-container-agent/spec"
)

const (
	executionEnvFargate = "AWS_ECS_FARGATE"
	executionEnvEC2     = "AWS_ECS_EC2"

	// ExecutionEnvExternal : (experimental) It is a definition that is handled internally, not an environment variable.
	ExecutionEnvExternal = "ECS_EXTERNAL"
)

var (
	logger                           = logging.GetLogger("ecs")
	taskMetadataWaitForReadyInterval = 3 * time.Second
)

// TaskMetadataEndpointClient interface gets task metadata and task stats
type TaskMetadataEndpointClient interface {
	TaskMetadataGetter
	TaskStatsGetter
}

type networkMode string

const (
	bridgeNetworkMode networkMode = "bridge"
	hostNetworkMode   networkMode = "host"
	awsvpcNetworkMode networkMode = "awsvpc"
)

type ecsPlatform struct {
	client      TaskMetadataEndpointClient
	provider    provider
	networkMode networkMode
}

// NewECSPlatform creates a new Platform
func NewECSPlatform(ctx context.Context, metadataURI string, executionEnv string, ignoreContainer *regexp.Regexp) (platform.Platform, error) {
	c, err := taskmetadata.NewClient(metadataURI, ignoreContainer)
	if err != nil {
		return nil, err
	}

	p, err := resolveProvider(executionEnv)
	if err != nil {
		return nil, err
	}

	meta, err := getTaskMetadata(ctx, c, taskMetadataWaitForReadyInterval)
	if err != nil {
		return nil, err
	}

	nm, err := detectNetworkMode(meta)
	if err != nil {
		return nil, err
	}

	return &ecsPlatform{
		client:      c,
		provider:    p,
		networkMode: nm,
	}, nil
}

// GetMetricGenerators gets metric generators
func (p *ecsPlatform) GetMetricGenerators() []metric.Generator {
	g := []metric.Generator{
		newMetricGenerator(p.client, hostinfo.NewGenerator()),
	}

	if p.networkMode != bridgeNetworkMode {
		g = append(g, metric.NewInterfaceGenerator())
	}

	return g
}

// GetSpecGenerators gets spec generator
func (p *ecsPlatform) GetSpecGenerators() []spec.Generator {
	return []spec.Generator{
		newSpecGenerator(p.client, p.provider),
		&spec.CPUGenerator{},
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
	return status == "RUNNING"
}

func resolveProvider(executionEnv string) (provider, error) {
	switch executionEnv {
	case executionEnvFargate:
		return fargateProvider, nil
	case executionEnvEC2:
		return ecsProvider, nil
	case ExecutionEnvExternal:
		return externalProvider, nil
	default:
		return provider("UNKNOWN"), fmt.Errorf("unknown execution env: %q", executionEnv)
	}
}

func detectNetworkMode(meta *ecsTypes.TaskResponse) (networkMode, error) {
	if len(meta.Containers) == 0 {
		return "", errors.New("there are no containers")
	}

	if len(meta.Containers[0].Networks) == 0 {
		return "", errors.New("there are no networks")
	}

	nm := meta.Containers[0].Networks[0].NetworkMode
	switch nm {
	case "default", "bridge":
		return bridgeNetworkMode, nil
	case "host":
		return hostNetworkMode, nil
	case "awsvpc":
		return awsvpcNetworkMode, nil
	default:
		return "", fmt.Errorf("unsupported NetworkMode: %v", nm)
	}
}

func getTaskMetadata(ctx context.Context, client TaskMetadataGetter, interval time.Duration) (*ecsTypes.TaskResponse, error) {
	// GetTaskMetadata will return an error until all containers associated with the task have been created.
	// To avoid exiting with this error, retry until GetTaskMetadata succeeds.
	for {
		meta, err := client.GetTaskMetadata(ctx)
		if err == nil {
			return meta, nil
		}

		logger.Infof("wait for the task API to be ready: %q", err)

		select {
		case <-time.After(interval):
			continue
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}
