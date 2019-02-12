package ecsawsvpc

import (
	"context"
	"strings"
	"time"

	ecsTypes "github.com/aws/amazon-ecs-agent/agent/handlers/v2"
	"github.com/aws/aws-sdk-go/aws/arn"

	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/platform"
	"github.com/mackerelio/mackerel-container-agent/platform/ecsawsvpc/taskmetadata"
	agentSpec "github.com/mackerelio/mackerel-container-agent/spec"
)

const resourcePrefix = "task/"

type specGenerator struct {
	client    taskmetadata.Client
	isFargate bool
}

func newSpecGenerator(client taskmetadata.Client, isFargate bool) *specGenerator {
	return &specGenerator{
		client:    client,
		isFargate: isFargate,
	}
}

// Generate generates metadata
func (g *specGenerator) Generate(ctx context.Context) (interface{}, error) {
	meta, err := g.client.GetMetadata(ctx)
	if err != nil {
		return nil, err
	}
	spec, err := generateSpec(meta)
	if err != nil {
		return nil, err
	}

	return &agentSpec.CloudHostname{
		Cloud: &mackerel.Cloud{
			Provider: resolvePlatform(g.isFargate),
			MetaData: spec,
		},
		Hostname: spec.Task,
	}, nil
}

func resolvePlatform(isFargate bool) string {
	if isFargate {
		return string(platform.Fargate)
	}
	return string(platform.ECS)
}

type taskSpec struct {
	Cluster            string      `json:"cluster,omitempty"`
	Task               string      `json:"task,omitempty"`
	TaskARN            string      `json:"task_arn,omitempty"`
	TaskFamily         string      `json:"task_family,omitempty"`
	TaskVersion        string      `json:"task_version,omitempty"`
	DesiredStatus      string      `json:"desired_status,omitempty"`
	KnownStatus        string      `json:"known_status,omitempty"`
	Containers         []container `json:"containers,omitempty"`
	PullStartedAt      *time.Time  `json:"pull_started_at,omitempty"`
	PullStoppedAt      *time.Time  `json:"pull_stopped_at,omitempty"`
	ExecutionStoppedAt *time.Time  `json:"execution_stopped_at,omitempty"`
	Limits             limits      `json:"limits,omitempty"`
}

type container struct {
	DockerID      string            `json:"docker_id,omitempty"`
	DockerName    string            `json:"docker_name,omitempty"`
	Name          string            `json:"name,omitempty"`
	Image         string            `json:"image,omitempty"`
	ImageID       string            `json:"image_id,omitempty"`
	Ports         []port            `json:"ports,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	DesiredStatus string            `json:"desired_status,omitempty"`
	KnownStatus   string            `json:"known_status,omitempty"`
	ExitCode      *int              `json:"exit_code,omitempty"`
	Limits        limits            `json:"limits,omitempty"`
	CreatedAt     *time.Time        `json:"crated_at,omitempty"`
	StartedAt     *time.Time        `json:"started_at,omitempty"`
	FinishedAt    *time.Time        `json:"finished_at,omitempty"`
	Type          string            `json:"type,omitempty"`
	Networks      []network         `json:"networks,omitempty"`
	Health        *health           `json:"health,omitempty"`
}

type limits struct {
	CPU    *float64 `json:"cpu,omitempty"`
	Memory *int64   `json:"memory,omitempty"`
}

type port struct {
	ContainerPort uint16 `json:"container_port,omitempty"`
	Protocol      string `json:"protocol,omitempty"`
	HostPort      uint16 `json:"host_port,omitempty"`
}

type network struct {
	NetworkMode   string   `json:"network_mode,omitempty"`
	IPv4Addresses []string `json:"ipv4_addresses,omitempty"`
	IPv6Addresses []string `json:"ipv6_addresses,omitempty"`
}

type health struct {
	Status   int32      `json:"status,omitempty"`
	Since    *time.Time `json:"status_since,omitempty"`
	ExitCode int        `json:"exit_code,omitempty"`
	Output   string     `json:"output,omitempty"`
}

func generateSpec(task *ecsTypes.TaskResponse) (*taskSpec, error) {
	taskARN, err := arn.Parse(task.TaskARN)
	if err != nil {
		return nil, err
	}

	spec := &taskSpec{
		Cluster:            task.Cluster,
		Task:               strings.TrimPrefix(taskARN.Resource, resourcePrefix),
		TaskARN:            task.TaskARN,
		TaskFamily:         task.Family,
		TaskVersion:        task.Revision,
		DesiredStatus:      task.DesiredStatus,
		KnownStatus:        task.KnownStatus,
		PullStartedAt:      task.PullStartedAt,
		PullStoppedAt:      task.PullStoppedAt,
		ExecutionStoppedAt: task.ExecutionStoppedAt,
	}

	if task.Containers != nil {
		containers := make([]container, len(task.Containers))
		spec.Containers = containers

		for i, c := range task.Containers {
			containers[i] = container{
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
				Limits: limits{
					CPU:    c.Limits.CPU,
					Memory: c.Limits.Memory,
				},
			}

			if c.Ports != nil {
				ports := make([]port, len(c.Ports))
				for j, p := range c.Ports {
					ports[j] = port{
						ContainerPort: p.ContainerPort,
						HostPort:      p.HostPort,
						Protocol:      p.Protocol,
					}
				}
				containers[i].Ports = ports
			}

			if c.Networks != nil {
				networks := make([]network, len(c.Networks))
				for j, n := range c.Networks {
					networks[j] = network{
						NetworkMode:   n.NetworkMode,
						IPv4Addresses: n.IPv4Addresses,
						IPv6Addresses: n.IPv6Addresses,
					}
				}
				containers[i].Networks = networks
			}

			if h := c.Health; h != nil {
				containers[i].Health = &health{
					ExitCode: h.ExitCode,
					Output:   h.Output,
					Since:    h.Since,
					Status:   int32(h.Status),
				}
			}
		}
	}

	if task.Limits != nil {
		spec.Limits = limits{
			CPU:    task.Limits.CPU,
			Memory: task.Limits.Memory,
		}
	}

	return spec, nil
}
