package ecsv3

import (
	"context"
	"runtime"
	"time"

	ecsTypes "github.com/aws/amazon-ecs-agent/agent/handlers/v2"
	dockerTypes "github.com/docker/docker/api/types"

	"github.com/mackerelio/go-osstat/memory"
	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/metric"
)

// TaskStatsGetter interface fetch ECS task stats
type TaskStatsGetter interface {
	GetTaskStats(context.Context) (map[string]*dockerTypes.StatsJSON, error)
}

type metricGenerator struct {
	client       TaskMetadataEndpointClient
	hostMemTotal *float64
	prevStats    map[string]*dockerTypes.StatsJSON
	prevTime     time.Time
}

func newMetricGenerator(client TaskMetadataEndpointClient) *metricGenerator {
	return &metricGenerator{
		client: client,
	}
}

// Generate generates metric values
func (g *metricGenerator) Generate(ctx context.Context) (metric.Values, error) {
	stats, err := g.client.GetTaskStats(ctx)
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

	meta, err := g.client.GetTaskMetadata(ctx)
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

		calculateInterfaceMetrics(name, prev, curr, timeDelta, metricValues)
	}

	g.prevStats = stats
	g.prevTime = now

	return metricValues, nil
}

// GetGraphDefs gets graph definitions
func (g *metricGenerator) GetGraphDefs(ctx context.Context) ([]*mackerel.GraphDefsParam, error) {
	return nil, nil
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

func calculateCPUMetrics(prev, curr *dockerTypes.StatsJSON, timeDelta time.Duration) float64 {
	// calculate used cpu cores. (1core == 100.0)
	return float64(curr.CPUStats.CPUUsage.TotalUsage-prev.CPUStats.CPUUsage.TotalUsage) / float64(timeDelta.Nanoseconds()) * 100
}

func calculateMemoryMetrics(stats *dockerTypes.StatsJSON) float64 {
	return float64(stats.MemoryStats.Usage - stats.MemoryStats.Stats["cache"])
}

func calculateInterfaceMetrics(name string, prev, curr *dockerTypes.StatsJSON, timeDelta time.Duration, metricValues metric.Values) {
	for ifn, pv := range prev.Networks {
		cv, ok := curr.Networks[ifn]
		if !ok {
			continue
		}
		prefix := "interface." + name + "-" + metric.SanitizeMetricKey(ifn)
		metricValues[prefix+".rxBytes.delta"] = float64(cv.RxBytes-pv.RxBytes) / timeDelta.Seconds()
		metricValues[prefix+".txBytes.delta"] = float64(cv.TxBytes-pv.TxBytes) / timeDelta.Seconds()
	}
}
