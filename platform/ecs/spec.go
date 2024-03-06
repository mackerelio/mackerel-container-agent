package ecs

import (
	"context"
	"path"

	"github.com/aws/aws-sdk-go/aws/arn"
	ecsTypes "github.com/mackerelio/mackerel-container-agent/internal/amazon-ecs-agent/agent/handlers/v2"

	mackerel "github.com/mackerelio/mackerel-client-go"

	agentSpec "github.com/mackerelio/mackerel-container-agent/spec"
)

// TaskMetadataGetter interface fetch ECS task metadata
type TaskMetadataGetter interface {
	GetTaskMetadata(context.Context) (*ecsTypes.TaskResponse, error)
}

type specGenerator struct {
	client   TaskMetadataGetter
	provider provider
}

func newSpecGenerator(client TaskMetadataGetter, provider provider) *specGenerator {
	return &specGenerator{
		client:   client,
		provider: provider,
	}
}

func (g *specGenerator) Generate(ctx context.Context) (any, error) {
	meta, err := g.client.GetTaskMetadata(ctx)
	if err != nil {
		return nil, err
	}

	spec, err := generateSpec(meta)
	if err != nil {
		return nil, err
	}

	return &agentSpec.CloudHostname{
		Cloud: &mackerel.Cloud{
			Provider: string(g.provider),
			MetaData: spec,
		},
		Hostname: spec.Task,
	}, nil
}

func generateSpec(meta *ecsTypes.TaskResponse) (*taskSpec, error) {
	taskID, err := getTaskID(meta.TaskARN)
	if err != nil {
		return nil, err
	}

	spec := &taskSpec{
		Cluster:            meta.Cluster,
		Task:               taskID,
		TaskARN:            meta.TaskARN,
		TaskFamily:         meta.Family,
		TaskVersion:        meta.Revision,
		DesiredStatus:      meta.DesiredStatus,
		KnownStatus:        meta.KnownStatus,
		PullStartedAt:      meta.PullStartedAt,
		PullStoppedAt:      meta.PullStoppedAt,
		ExecutionStoppedAt: meta.ExecutionStoppedAt,
	}

	if meta.Containers != nil {
		containers := make([]containerSpec, len(meta.Containers))
		spec.Containers = containers

		for i, c := range meta.Containers {
			containers[i] = containerSpec{
				DockerID:      c.ID,
				DockerName:    c.DockerName,
				Name:          c.Name,
				Image:         c.Image,
				ImageID:       c.ImageID,
				Labels:        c.Labels,
				DesiredStatus: c.DesiredStatus,
				KnownStatus:   c.KnownStatus,
				ExitCode:      c.ExitCode,
				CreatedAt:     c.CreatedAt,
				StartedAt:     c.StartedAt,
				FinishedAt:    c.FinishedAt,
				Type:          c.Type,
				Limits: limitSpec{
					CPU:    c.Limits.CPU,
					Memory: c.Limits.Memory,
				},
			}

			if c.Ports != nil {
				ports := make([]portSpec, len(c.Ports))
				for j, p := range c.Ports {
					ports[j] = portSpec{
						ContainerPort: p.ContainerPort,
						HostPort:      p.HostPort,
						Protocol:      p.Protocol,
					}
				}
				containers[i].Ports = ports
			}

			if c.Networks != nil {
				networks := make([]networkSpec, len(c.Networks))
				for j, n := range c.Networks {
					networks[j] = networkSpec{
						NetworkMode:   n.NetworkMode,
						IPv4Addresses: n.IPv4Addresses,
						IPv6Addresses: n.IPv6Addresses,
					}
				}
				containers[i].Networks = networks
			}

			if h := c.Health; h != nil {
				containers[i].Health = &healthStatus{
					ExitCode: h.ExitCode,
					Output:   h.Output,
					Since:    h.Since,
					Status:   int32(h.Status),
				}
			}
		}
	}

	if meta.Limits != nil {
		spec.Limits = limitSpec{
			CPU:    meta.Limits.CPU,
			Memory: meta.Limits.Memory,
		}
	}

	return spec, nil
}

func getTaskID(taskARN string) (string, error) {
	a, err := arn.Parse(taskARN)
	if err != nil {
		return "", err
	}
	return path.Base(a.Resource), nil
}
