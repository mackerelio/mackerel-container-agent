package ecs

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"

	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/platform"
	"github.com/mackerelio/mackerel-container-agent/platform/ecs/task"
	agentSpec "github.com/mackerelio/mackerel-container-agent/spec"
)

const resourcePrefix = "task/"

type ecsSpec struct {
	Cluster       string         `json:"cluster,omitempty"`
	Task          string         `json:"task,omitempty"`
	TaskARN       string         `json:"task_arn,omitempty"`
	TaskFamily    string         `json:"task_family,omitempty"`
	TaskVersion   string         `json:"task_version,omitempty"`
	DesiredStatus string         `json:"desired_status,omitempty"`
	KnownStatus   string         `json:"known_status,omitempty"`
	Containers    []container    `json:"containers,omitempty"`
	Limits        resourceLimits `json:"limits,omitempty"`
}

type container struct {
	DockerID   string `json:"docker_id,omitempty"`
	DockerName string `json:"docker_name,omitempty"`
	Name       string `json:"name,omitempty"`
}

type resourceLimits struct {
	CPU    float64 `json:"cpu,omitempty"`
	Memory uint64  `json:"memory,omitempty"`
}

type specGenerator struct {
	task task.Task
}

func newSpecGenerator(task task.Task) *specGenerator {
	return &specGenerator{
		task: task,
	}
}

// Generate generates metadata
func (g *specGenerator) Generate(ctx context.Context) (interface{}, error) {
	meta, err := g.task.Metadata(ctx)
	if err != nil {
		return nil, err
	}

	taskARN, err := arn.Parse(meta.Arn)
	if err != nil {
		return nil, err
	}

	containers := make([]container, len(meta.Containers))
	for i, c := range meta.Containers {
		containers[i] = container{
			DockerID:   c.DockerID,
			DockerName: c.DockerName,
			Name:       c.Name,
		}
	}

	spec := &ecsSpec{
		Cluster:       meta.Instance.Cluster,
		Task:          strings.TrimPrefix(taskARN.Resource, resourcePrefix),
		TaskARN:       meta.Arn,
		TaskFamily:    meta.Family,
		TaskVersion:   meta.Version,
		DesiredStatus: meta.DesiredStatus,
		KnownStatus:   meta.KnownStatus,
		Containers:    containers,
		Limits: resourceLimits{
			CPU:    meta.Limits.CPU,
			Memory: meta.Limits.Memory,
		},
	}

	return &agentSpec.CloudHostname{
		Cloud: &mackerel.Cloud{
			Provider: string(platform.ECS),
			MetaData: spec,
		},
		Hostname: spec.Task,
	}, nil
}
