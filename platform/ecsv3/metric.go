package ecsv3

import (
	"context"

	dockerTypes "github.com/docker/docker/api/types"
	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/metric"
)

// TaskStatsGetter interface fetch ECS task stats
type TaskStatsGetter interface {
	GetTaskStats(context.Context) (map[string]*dockerTypes.StatsJSON, error)
}

type metricGenerator struct {
	client TaskStatsGetter
}

func newMetricGenerator(client TaskStatsGetter) *metricGenerator {
	return &metricGenerator{
		client: client,
	}
}

// Generate generates metric values
func (g *metricGenerator) Generate(ctx context.Context) (metric.Values, error) {
	return nil, nil
}

// GetGraphDefs gets graph definitions
func (g *metricGenerator) GetGraphDefs(ctx context.Context) ([]*mackerel.GraphDefsParam, error) {
	return nil, nil
}
