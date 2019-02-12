package ecs

import (
	"context"
	"time"

	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/metric"
	"github.com/mackerelio/mackerel-container-agent/platform/ecs/task"
)

type metricGenerator struct {
	task      task.Task
	prevStats map[string]*task.Stats
	prevTime  time.Time
}

func newMetricGenerator(task task.Task) *metricGenerator {
	return &metricGenerator{
		task: task,
	}
}

// Generate generates metric values
func (g *metricGenerator) Generate(ctx context.Context) (metric.Values, error) {
	stats, err := g.task.Stats(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()

	if g.prevStats == nil || g.prevTime.Before(now.Add(-10*time.Minute)) {
		g.prevStats = stats
		g.prevTime = now
		return nil, nil
	}

	var metricValues = make(metric.Values)

	task, err := g.task.Metadata(ctx)
	if err != nil {
		return nil, err
	}

	timeDelta := now.Sub(g.prevTime)
	for _, c := range task.Containers {
		prev, ok := g.prevStats[c.Name]
		if !ok || prev == nil {
			continue
		}
		curr, ok := stats[c.Name]
		if !ok || curr == nil {
			continue
		}

		name := metric.SanitizeMetricKey(c.Name)

		metricValues["container.cpu."+name+".usage"] = calculateCPUMetrics(prev, curr, timeDelta)
		metricValues["container.memory."+name+".usage"] = calculateMemoryMetrics(curr)

		var cpuLimit float64
		if curr.CPU.Limit > 0.0 {
			cpuLimit = curr.CPU.Limit
		} else {
			cpuLimit = task.Limits.CPU
		}
		metricValues["container.cpu."+name+".limit"] = cpuLimit

		var memoryLimit float64
		if curr.Memory.Limit > 0 {
			memoryLimit = float64(curr.Memory.Limit)
		} else {
			memoryLimit = float64(task.Limits.Memory)
		}
		metricValues["container.memory."+name+".limit"] = memoryLimit

		if g.task.IsPrivateNetworkMode() {
			generateInterfaceMetrics(name, prev, curr, timeDelta, metricValues)
		}
	}

	g.prevStats = stats
	g.prevTime = now

	return metricValues, nil
}

func calculateCPUMetrics(prev, curr *task.Stats, timeDelta time.Duration) float64 {
	// calculate used cpu cores. (1core == 100.0)
	return float64(curr.CPU.Total-prev.CPU.Total) / float64(timeDelta.Nanoseconds()) * 100.0
}

func calculateMemoryMetrics(stats *task.Stats) float64 {
	return float64(stats.Memory.Usage - stats.Memory.Stats["cache"])
}

func generateInterfaceMetrics(name string, prev, curr *task.Stats, timeDelta time.Duration, metricValues metric.Values) {
	for ifName, prevValue := range prev.Network {
		currValue, ok := curr.Network[ifName]
		if !ok {
			continue
		}
		prefix := "interface." + name + "-" + metric.SanitizeMetricKey(ifName)
		metricValues[prefix+".rxBytes.delta"] = float64(currValue.RxBytes-prevValue.RxBytes) / timeDelta.Seconds()
		metricValues[prefix+".txBytes.delta"] = float64(currValue.TxBytes-prevValue.TxBytes) / timeDelta.Seconds()
	}
}

// GetGraphDefs gets graph definitions
func (g *metricGenerator) GetGraphDefs(context.Context) ([]*mackerel.GraphDefsParam, error) {
	return nil, nil
}
