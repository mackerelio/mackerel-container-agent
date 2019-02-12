package ecsawsvpc

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"

	ecsTypes "github.com/aws/amazon-ecs-agent/agent/handlers/v2"
	dockerTypes "github.com/docker/docker/api/types"

	"github.com/mackerelio/mackerel-container-agent/metric"
	"github.com/mackerelio/mackerel-container-agent/platform/ecsawsvpc/taskmetadata"
)

func TestGenerateStats(t *testing.T) {
	ctx := context.Background()
	client := taskmetadata.NewMockClient(
		taskmetadata.MockGetMetadata(func(context.Context) (*ecsTypes.TaskResponse, error) {
			raw, err := ioutil.ReadFile("taskmetadata/testdata/metadata.json")
			if err != nil {
				return nil, err
			}
			var task ecsTypes.TaskResponse
			err = json.Unmarshal(raw, &task)
			if err != nil {
				return nil, err
			}
			return &task, nil
		}),
		taskmetadata.MockGetStats(func(context.Context) (map[string]*dockerTypes.Stats, error) {
			raw, err := ioutil.ReadFile("taskmetadata/testdata/stats.json")
			if err != nil {
				return nil, err
			}
			var stats map[string]*dockerTypes.Stats
			err = json.Unmarshal(raw, &stats)
			if err != nil {
				return nil, err
			}
			return stats, nil
		}),
	)
	generator := newMetricGenerator(client)
	_, err := generator.Generate(ctx) // Store metrics to generator.prevStats.
	if err != nil {
		t.Errorf("Generate() should not raise error: %v", err)
	}
	got, err := generator.Generate(ctx)
	if err != nil {
		t.Errorf("Generate() should not raise error: %v", err)
	}
	expected := metric.Values{
		"container.cpu.nginx-curl.usage":             0,    // Rsult is 0 because use the same data.
		"container.cpu._internal_ecs_pause.usage":    0,    // Rsult is 0 because use the same data.
		"container.cpu.nginx-curl.limit":             50.0, // Task CPU Limit. (Because contaienr CPU Limit means `cpu.shares`)
		"container.cpu._internal_ecs_pause.limit":    50.0, // Task CPU Limit.
		"container.memory.nginx-curl.usage":          5.812224e+06,
		"container.memory._internal_ecs_pause.usage": 2.351104e+06,
		"container.memory.nginx-curl.limit":          256.0 * MiB,
		"container.memory._internal_ecs_pause.limit": 512.0 * MiB,
	}
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("Generate() expected %v, got %v", expected, got)
	}
}

func TestGetMemoryLimit(t *testing.T) {
	generator := &metricGenerator{
		hostMemTotal: func(f float64) *float64 { return &f }(float64(3 * MiB)),
	}
	tests := []struct {
		container *ecsTypes.ContainerResponse // contaienr limit
		meta      *ecsTypes.TaskResponse      // task limit
		generator *metricGenerator
		expected  float64
	}{
		{
			container: &ecsTypes.ContainerResponse{}, // no limit
			meta:      &ecsTypes.TaskResponse{},      // no limit
			generator: generator,
			expected:  float64(3 * MiB),
		},
		{
			container: &ecsTypes.ContainerResponse{}, // no limit
			meta: &ecsTypes.TaskResponse{
				Limits: &ecsTypes.LimitsResponse{
					Memory: func(i int64) *int64 { return &i }(2),
				},
			},
			generator: generator,
			expected:  float64(2 * MiB),
		},
		{
			container: &ecsTypes.ContainerResponse{
				Limits: ecsTypes.LimitsResponse{
					Memory: func(i int64) *int64 { return &i }(1),
				},
			},
			meta: &ecsTypes.TaskResponse{
				Limits: &ecsTypes.LimitsResponse{
					Memory: func(i int64) *int64 { return &i }(2),
				},
			},
			generator: generator,
			expected:  float64(1 * MiB),
		},
	}

	for _, tc := range tests {
		got := tc.generator.getMemoryLimit(tc.container, tc.meta)
		if got != tc.expected {
			t.Errorf("getMemoryLimit() expected %.1f, got %.1f", tc.expected, got)
		}
	}
}
