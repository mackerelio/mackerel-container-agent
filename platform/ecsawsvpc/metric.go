package ecsawsvpc

import (
	"context"
	"runtime"
	"time"

	ecsTypes "github.com/aws/amazon-ecs-agent/agent/handlers/v2"
	dockerTypes "github.com/docker/docker/api/types"

	"github.com/mackerelio/go-osstat/memory"
	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/metric"
	"github.com/mackerelio/mackerel-container-agent/platform/ecsawsvpc/taskmetadata"
)

type metricGenerator struct {
	client       taskmetadata.Client
	hostMemTotal *float64
	prevStats    map[string]*dockerTypes.Stats
	prevTime     time.Time
}

func newMetricGenerator(client taskmetadata.Client) *metricGenerator {
	return &metricGenerator{
		client: client,
	}
}

// Generate generates metric values
func (g *metricGenerator) Generate(ctx context.Context) (metric.Values, error) {
	stats, err := g.client.GetStats(ctx)
	if err != nil {
		return nil, err
	}
	if g.hostMemTotal == nil {
		memory, err := memory.Get()
		if err != nil {
			return nil, err
		}
		total := float64(memory.Total)
		g.hostMemTotal = &total
	}
	now := time.Now()
	if g.prevStats == nil || g.prevTime.Before(now.Add(-10*time.Minute)) {
		g.prevStats = stats
		g.prevTime = now
		return nil, nil
	}

	meta, err := g.client.GetMetadata(ctx)
	if err != nil {
		return nil, err
	}

	timeDelta := now.Sub(g.prevTime)
	metricValues := make(metric.Values)
	for _, c := range meta.Containers {
		prev, ok := g.prevStats[c.ID]
		if !ok || prev == nil { // stats of the volume container can be nil value.
			continue
		}
		curr, ok := stats[c.ID]
		if !ok || curr == nil { // stats of the volume container can be nil value.
			continue
		}
		name := metric.SanitizeMetricKey(c.Name)
		metricValues["container.cpu."+name+".usage"] = calculateCPUMetrics(prev, curr, timeDelta)
		metricValues["container.cpu."+name+".limit"] = getCPULimit(meta)
		metricValues["container.memory."+name+".usage"] = calculateMemoryMetrics(curr)
		metricValues["container.memory."+name+".limit"] = g.getMemoryLimit(&c, meta)
	}

	g.prevStats = stats
	g.prevTime = now

	return metricValues, nil
}

func (g *metricGenerator) getMemoryLimit(c *ecsTypes.ContainerResponse, meta *ecsTypes.TaskResponse) float64 {
	if c.Limits.Memory != nil && *c.Limits.Memory != 0 {
		return float64(*c.Limits.Memory * MiB)
	} else if meta.Limits != nil && meta.Limits.Memory != nil && *meta.Limits.Memory != 0 {
		return float64(*meta.Limits.Memory * MiB)
	}
	return *g.hostMemTotal
}

func getCPULimit(meta *ecsTypes.TaskResponse) float64 {
	// Return Task CPU Limit or Host CPU Limit because Container CPU Limit means `cpu.shares`.
	if meta.Limits != nil && meta.Limits.CPU != nil && *meta.Limits.CPU != 0.0 {
		return *meta.Limits.CPU * 100
	}
	return float64(runtime.NumCPU() * 100)
}

func calculateCPUMetrics(prev, curr *dockerTypes.Stats, timeDelta time.Duration) float64 {
	// calculate used cpu cores. (1core == 100.0)
	return float64(curr.CPUStats.CPUUsage.TotalUsage-prev.CPUStats.CPUUsage.TotalUsage) / float64(timeDelta.Nanoseconds()) * 100
}

func calculateMemoryMetrics(stats *dockerTypes.Stats) float64 {
	return float64(stats.MemoryStats.Usage - stats.MemoryStats.Stats["cache"])
}

// GetGraphDefs gets graph definitions
func (g *metricGenerator) GetGraphDefs(context.Context) ([]*mackerel.GraphDefsParam, error) {
	return nil, nil
}
