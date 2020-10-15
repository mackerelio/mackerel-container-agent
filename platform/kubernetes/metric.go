package kubernetes

import (
	"context"
	"fmt"
	"strconv"
	"time"

	kubeletTypes "github.com/mackerelio/mackerel-container-agent/internal/k8s-apis/stats/v1alpha1"
	kubernetesTypes "k8s.io/api/core/v1"

	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/metric"
	"github.com/mackerelio/mackerel-container-agent/platform/kubernetes/kubelet"
	"github.com/mackerelio/mackerel-container-agent/platform/kubernetes/kubernetesapi"
)

type metricGenerator struct {
	client       kubelet.Client
	apiClient    kubernetesapi.Client
	hostMemTotal *float64
	hostNumCores *float64
	prevStats    *kubeletTypes.PodStats
	prevTime     time.Time
}

func newMetricGenerator(client kubelet.Client, apiClient kubernetesapi.Client) *metricGenerator {
	return &metricGenerator{
		client:    client,
		apiClient: apiClient,
	}
}

// getHostCapacityFromKubelet may not work on newer Kubernetes (1.18+), due to API deprecation.
func (g *metricGenerator) getHostCapacityFromKubelet(ctx context.Context) (numCores, memTotal float64, err error) {
	machineInfo, err := g.client.GetSpec(ctx)
	if err != nil {
		return 0, 0, err
	}
	return float64(machineInfo.NumCores), float64(machineInfo.MemoryCapacity), nil
}

func (g *metricGenerator) getHostCapacityFromKubernetesAPI(ctx context.Context) (numCores, memTotal float64, err error) {
	nodeInfo, err := g.apiClient.GetNode(ctx)
	if err != nil {
		return 0, 0, err
	}

	cores := nodeInfo.Status.Capacity.Cpu()
	coresInt, ok := cores.AsInt64()
	if !ok {
		return 0, 0, fmt.Errorf("failed to parse CPU cores %s", cores)
	}

	mem := nodeInfo.Status.Capacity.Memory()
	memInt, ok := mem.AsInt64()
	if !ok {
		return 0, 0, fmt.Errorf("failed to parse memory %s", mem)
	}

	return float64(coresInt), float64(memInt), nil
}

func (g *metricGenerator) Generate(ctx context.Context) (metric.Values, error) {
	stats, err := g.client.GetPodStats(ctx)
	if err != nil {
		return nil, err
	}
	if g.hostMemTotal == nil || g.hostNumCores == nil {
		var cores, total float64
		cores, total, err = g.getHostCapacityFromKubelet(ctx)
		if err == kubelet.ErrNotFound {
			cores, total, err = g.getHostCapacityFromKubernetesAPI(ctx)
			if err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}
		if g.hostMemTotal == nil {
			g.hostMemTotal = &total
		}
		if g.hostNumCores == nil {
			g.hostNumCores = &cores
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
				metrics["container.memory."+name+".usage"] = float64(*currContainer.Memory.UsageBytes)
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
