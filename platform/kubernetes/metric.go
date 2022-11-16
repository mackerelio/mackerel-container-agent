package kubernetes

import (
	"context"
	"strconv"
	"time"

	kubernetesTypes "k8s.io/api/core/v1"
	kubeletTypes "k8s.io/kubelet/pkg/apis/stats/v1alpha1"

	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/metric"
	"github.com/mackerelio/mackerel-container-agent/metric/hostinfo"
	"github.com/mackerelio/mackerel-container-agent/platform/kubernetes/kubelet"
)

type metricGenerator struct {
	client            kubelet.Client
	hostInfoGenerator hostinfo.Generator
	hostMemTotal      *float64
	hostNumCores      *float64
	prevStats         *kubeletTypes.PodStats
	prevTime          time.Time
}

func newMetricGenerator(client kubelet.Client, hostinfoGenerator hostinfo.Generator) *metricGenerator {
	return &metricGenerator{
		client:            client,
		hostInfoGenerator: hostinfoGenerator,
	}
}

func (g *metricGenerator) Generate(ctx context.Context) (metric.Values, error) {
	stats, err := g.client.GetPodStats(ctx)
	if err != nil {
		return nil, err
	}
	if g.hostMemTotal == nil || g.hostNumCores == nil {
		memTotal, cpuCores, err := g.hostInfoGenerator.Generate()
		if err != nil {
			return nil, err
		}
		if g.hostMemTotal == nil {
			g.hostMemTotal = &memTotal
		}
		if g.hostNumCores == nil {
			g.hostNumCores = &cpuCores
		}
	}

	now := time.Now()
	if g.prevStats == nil || g.prevTime.Before(now.Add(-10*time.Minute)) {
		g.prevStats = stats
		g.prevTime = now
		return nil, nil
	}

	pod, err := g.client.GetPod(ctx)
	if err != nil {
		return nil, err
	}

	delta := now.Sub(g.prevTime)
	metrics := make(metric.Values)
	for _, prevContainer := range g.prevStats.Containers {
		for _, currContainer := range stats.Containers {
			if currContainer.Name == prevContainer.Name {
				name := metric.SanitizeMetricKey(currContainer.Name)
				metrics["container.cpu."+name+".usage"] = calculateCPUMetrics(&prevContainer, &currContainer, delta)
				if currContainer.Memory.WorkingSetBytes != nil {
					metrics["container.memory."+name+".usage"] = float64(*currContainer.Memory.WorkingSetBytes)
				}
				for _, c := range pod.Spec.Containers {
					if c.Name == currContainer.Name {
						metrics["container.cpu."+name+".limit"] = g.getCPULimit(&c)
						metrics["container.memory."+name+".limit"] = g.getMermoryLimit(&c)
						break
					}
				}
			}
		}
	}

	g.prevStats = stats
	g.prevTime = now

	return metrics, nil
}

func (g *metricGenerator) getMermoryLimit(container *kubernetesTypes.Container) float64 {
	limit := *g.hostMemTotal
	if v, ok := container.Resources.Limits["memory"]; ok && v.Format != "" {
		i, _ := v.AsInt64()
		limit = float64(i)
	}
	return limit
}

func (g *metricGenerator) getCPULimit(container *kubernetesTypes.Container) float64 {
	limit := *g.hostNumCores * 100
	if v, ok := container.Resources.Limits["cpu"]; ok {
		if d := v.AsDec(); d != nil {
			if v, err := strconv.ParseFloat(d.String(), 64); err == nil {
				limit = v * 100
			}
		}
	}
	return limit
}

func calculateCPUMetrics(prev, curr *kubeletTypes.ContainerStats, delta time.Duration) float64 {
	return float64(*curr.CPU.UsageCoreNanoSeconds-*prev.CPU.UsageCoreNanoSeconds) / float64(delta.Nanoseconds()) * 100
}

func (g *metricGenerator) GetGraphDefs(context.Context) ([]*mackerel.GraphDefsParam, error) {
	return nil, nil
}
