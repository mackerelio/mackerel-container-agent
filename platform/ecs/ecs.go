package ecs

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/mackerelio/golib/logging"

	"github.com/mackerelio/mackerel-container-agent/metric"
	"github.com/mackerelio/mackerel-container-agent/platform"
	"github.com/mackerelio/mackerel-container-agent/platform/ecs/agent"
	"github.com/mackerelio/mackerel-container-agent/platform/ecs/cgroupfs"
	"github.com/mackerelio/mackerel-container-agent/platform/ecs/docker"
	"github.com/mackerelio/mackerel-container-agent/platform/ecs/instance"
	"github.com/mackerelio/mackerel-container-agent/platform/ecs/procfs"
	"github.com/mackerelio/mackerel-container-agent/platform/ecs/task"
	"github.com/mackerelio/mackerel-container-agent/spec"
)

var logger = logging.GetLogger("ecs")

var procfsMountPoint = procfs.DefaultMountPoint
var dockerSockMountPoint = "/var/run/docker.sock"
var cgroupfsMountPoint = "/host/sys/fs/cgroup"

type ecsPlatform struct {
	task task.Task
}

// NewECSPlatform creates a new Platform
func NewECSPlatform(ctx context.Context, instanceClient instance.Client, ignoreContainer *regexp.Regexp) (platform.Platform, error) {
	task, err := newTask(ctx, instanceClient, ignoreContainer)
	if err != nil {
		return nil, err
	}
	return &ecsPlatform{
		task: task,
	}, nil
}

// GetMetricGenerators gets metric generators
func (p *ecsPlatform) GetMetricGenerators() []metric.Generator {
	var generator = []metric.Generator{
		newMetricGenerator(p.task),
	}
	if !p.task.IsPrivateNetworkMode() {
		generator = append(generator, metric.NewInterfaceGenerator())
	}
	return generator
}

// GetSpecGenerators gets spec generator
func (p *ecsPlatform) GetSpecGenerators() []spec.Generator {
	return []spec.Generator{
		newSpecGenerator(p.task),
	}
}

// GetCustomIdentifier gets custom identifier
func (p *ecsPlatform) GetCustomIdentifier(context.Context) (string, error) {
	return "", nil
}

// StatusRunning reports p status is running
func (p *ecsPlatform) StatusRunning(ctx context.Context) bool {
	meta, err := p.task.Metadata(ctx)
	if err != nil {
		logger.Warningf("failed to get metadata: %s", err)
		return false
	}
	return strings.EqualFold("running", meta.KnownStatus)
}

func newTask(ctx context.Context, instanceClient instance.Client, ignoreContainer *regexp.Regexp) (task.Task, error) {
	proc, err := procfs.Self(procfsMountPoint)
	if err != nil {
		return nil, err
	}
	dockerClient, err := docker.NewClient(dockerSockMountPoint)
	if err != nil {
		return nil, err
	}
	agentClient, err := newAgentClient(ctx, instanceClient)
	if err != nil {
		return nil, err
	}
	cgroup, err := cgroupfs.NewCgroup(cgroupfsMountPoint)
	if err != nil {
		return nil, err
	}
	return task.NewTaskWithProc(ctx, proc, dockerClient, agentClient, cgroup, ignoreContainer)
}

func newAgentClient(ctx context.Context, instanceClient instance.Client) (agent.Client, error) {
	host, err := instanceClient.GetLocalIPv4(ctx)
	if err != nil {
		return nil, err
	}
	return agent.NewClient(fmt.Sprintf("http://%s:%d", host, agent.DefaultPort))
}
