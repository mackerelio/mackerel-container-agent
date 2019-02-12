package ecs

import (
	"context"
	"reflect"
	"runtime"
	"testing"

	"github.com/mackerelio/mackerel-container-agent/metric"
	"github.com/mackerelio/mackerel-container-agent/platform/ecs/task"
)

func TestGenerateMetric(t *testing.T) {
	mockTask := task.NewMockTask(
		task.MockStats(
			func(context.Context) (map[string]*task.Stats, error) {
				return map[string]*task.Stats{
					"nginx": &task.Stats{
						CPU: task.CPUStats{
							Total: 413514641021,
						},
						Memory: task.MemoryStats{
							Usage: 9666560,
							Stats: map[string]uint64{
								"cache": 8192,
								"rss":   1511424,
								"swap":  0,
							},
						},
						Network: map[string]task.NetworkStats{
							"eth0": {
								Name:    "eth0",
								RxBytes: 26160,
								TxBytes: 38896,
							},
						},
					},
					"mackerel-container-agent": &task.Stats{
						CPU: task.CPUStats{
							Total: 35068405,
							Limit: 25.0,
						},
						Memory: task.MemoryStats{
							Usage: 46956544,
							Stats: map[string]uint64{
								"cache": 12926976,
								"rss":   30441472,
								"swap":  0,
							},
							Limit: 134217728,
						},
						Network: map[string]task.NetworkStats{
							"eth0": {
								Name:    "eth0",
								RxBytes: 670818,
								TxBytes: 1229820,
							},
						},
					},
				}, nil
			},
		),
		task.MockMetadata(
			func(context.Context) (*task.Metadata, error) {
				return &task.Metadata{
					Limits: task.ResourceLimits{
						CPU:    float64(runtime.NumCPU()) * 100.0,
						Memory: 268435456,
					},
					Containers: []task.Container{
						{Name: "nginx"},
						{Name: "mackerel-container-agent"},
					},
				}, nil
			},
		),
		task.MockIsPrivateNetworkMode(
			func() bool {
				return true
			},
		),
	)

	generator := newMetricGenerator(mockTask)
	ctx := context.Background()

	_, err := generator.Generate(ctx)
	if err != nil {
		t.Errorf("Generate() should not raise error: %v", err)
	}
	got, err := generator.Generate(ctx)
	if err != nil {
		t.Errorf("Generate() should not raise error: %v", err)
	}

	expected := metric.Values{
		"container.cpu.nginx.usage":                             0, // Rsult is 0 because use the same data.
		"container.cpu.nginx.limit":                             float64(runtime.NumCPU() * 100.0),
		"container.cpu.mackerel-container-agent.usage":          0, // Rsult is 0 because use the same data.
		"container.cpu.mackerel-container-agent.limit":          25.0,
		"container.memory.nginx.usage":                          9666560 - 8192,
		"container.memory.nginx.limit":                          268435456,
		"container.memory.mackerel-container-agent.usage":       46956544 - 12926976,
		"container.memory.mackerel-container-agent.limit":       134217728,
		"interface.nginx-eth0.rxBytes.delta":                    0, // Rsult is 0 because use the same data.
		"interface.nginx-eth0.txBytes.delta":                    0, // Rsult is 0 because use the same data.
		"interface.mackerel-container-agent-eth0.rxBytes.delta": 0, // Rsult is 0 because use the same data.
		"interface.mackerel-container-agent-eth0.txBytes.delta": 0, // Rsult is 0 because use the same data.
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Generate() expected %v, got %v", expected, got)
	}
}
