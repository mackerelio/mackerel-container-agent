package task

import ecsTypes "github.com/aws/amazon-ecs-agent/agent/handlers/v1"

// Metadata represents task metadata
type Metadata struct {
	Arn           string
	DesiredStatus string
	KnownStatus   string
	Family        string
	Version       string
	Containers    []Container `json:"Containers"`

	Instance *ecsTypes.MetadataResponse
	Limits   ResourceLimits
}

// Container represents container
type Container struct {
	DockerID   string
	DockerName string
	Name       string
}

// ResourceLimits represents task resouce limits
type ResourceLimits struct {
	CPU    float64
	Memory uint64
}

// Stats represents resource stats
type Stats struct {
	CPU     CPUStats
	Memory  MemoryStats
	Network map[string]NetworkStats
}

// CPUStats represents cpu stats
type CPUStats struct {
	Total uint64
	Limit float64
}

// MemoryStats represents memory stats
type MemoryStats struct {
	Usage uint64
	Stats map[string]uint64
	Limit uint64
}

// NetworkStats represents network stats
type NetworkStats struct {
	Name             string
	RxBytes, TxBytes uint64
}
