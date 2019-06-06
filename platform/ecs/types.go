package ecs

import "time"

type taskSpec struct {
	Cluster            string          `json:"cluster,omitempty"`
	Task               string          `json:"task,omitempty"`
	TaskARN            string          `json:"task_arn,omitempty"`
	TaskFamily         string          `json:"task_family,omitempty"`
	TaskVersion        string          `json:"task_version,omitempty"`
	DesiredStatus      string          `json:"desired_status,omitempty"`
	KnownStatus        string          `json:"known_status,omitempty"`
	Containers         []containerSpec `json:"containers,omitempty"`
	PullStartedAt      *time.Time      `json:"pull_started_at,omitempty"`
	PullStoppedAt      *time.Time      `json:"pull_stopped_at,omitempty"`
	ExecutionStoppedAt *time.Time      `json:"execution_stopped_at,omitempty"`
	Limits             limitSpec       `json:"limits,omitempty"`
}

type containerSpec struct {
	DockerID      string            `json:"docker_id,omitempty"`
	DockerName    string            `json:"docker_name,omitempty"`
	Name          string            `json:"name,omitempty"`
	Image         string            `json:"image,omitempty"`
	ImageID       string            `json:"image_id,omitempty"`
	Ports         []portSpec        `json:"ports,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	DesiredStatus string            `json:"desired_status,omitempty"`
	KnownStatus   string            `json:"known_status,omitempty"`
	ExitCode      *int              `json:"exit_code,omitempty"`
	Limits        limitSpec         `json:"limits,omitempty"`
	CreatedAt     *time.Time        `json:"crated_at,omitempty"`
	StartedAt     *time.Time        `json:"started_at,omitempty"`
	FinishedAt    *time.Time        `json:"finished_at,omitempty"`
	Type          string            `json:"type,omitempty"`
	Networks      []networkSpec     `json:"networks,omitempty"`
	Health        *healthStatus     `json:"health,omitempty"`
}

type limitSpec struct {
	CPU    *float64 `json:"cpu,omitempty"`
	Memory *int64   `json:"memory,omitempty"`
}

type portSpec struct {
	ContainerPort uint16 `json:"container_port,omitempty"`
	Protocol      string `json:"protocol,omitempty"`
	HostPort      uint16 `json:"host_port,omitempty"`
}

type networkSpec struct {
	NetworkMode   string   `json:"network_mode,omitempty"`
	IPv4Addresses []string `json:"ipv4_addresses,omitempty"`
	IPv6Addresses []string `json:"ipv6_addresses,omitempty"`
}

type healthStatus struct {
	Status   int32      `json:"status,omitempty"`
	Since    *time.Time `json:"status_since,omitempty"`
	ExitCode int        `json:"exit_code,omitempty"`
	Output   string     `json:"output,omitempty"`
}
