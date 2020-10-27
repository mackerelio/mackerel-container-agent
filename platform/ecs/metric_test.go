package ecs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	ecsTypes "github.com/aws/amazon-ecs-agent/agent/handlers/v2"
	dockerTypes "github.com/docker/docker/api/types"

	"github.com/mackerelio/mackerel-container-agent/metric"
	"github.com/mackerelio/mackerel-container-agent/metric/hostinfo"
)

type mockTaskMetadataEndpointClient struct {
	metadataPath string
	statsPath    string
}

func (m *mockTaskMetadataEndpointClient) GetTaskMetadata(ctx context.Context) (*ecsTypes.TaskResponse, error) {
	f, err := os.Open(m.metadataPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var res ecsTypes.TaskResponse
	if err := json.NewDecoder(f).Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (m *mockTaskMetadataEndpointClient) GetTaskStats(ctx context.Context) (map[string]*dockerTypes.StatsJSON, error) {
	f, err := os.Open(m.statsPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var res map[string]*dockerTypes.StatsJSON
	if err := json.NewDecoder(f).Decode(&res); err != nil {
		return nil, err
	}

	return res, nil
}

func TestGenerateMetric(t *testing.T) {
	tests := []struct {
		mode     string
		expected metric.Values
	}{
		{
			"ec2_bridge",
			metric.Values{
				"container.cpu.mackerel-container-agent.usage":          0.0, // Result is 0 because use the same data.
				"container.cpu.mackerel-container-agent.limit":          25.0,
				"container.memory.mackerel-container-agent.usage":       1.2111872e+07,
				"container.memory.mackerel-container-agent.limit":       134217728.0, // 128MiB
				"interface.mackerel-container-agent-eth0.rxBytes.delta": 0,
				"interface.mackerel-container-agent-eth0.txBytes.delta": 0,
			},
		},
		{
			"ec2_host",
			metric.Values{
				"container.cpu.mackerel-container-agent.usage":    0.0, // Result is 0 because use the same data.
				"container.cpu.mackerel-container-agent.limit":    25.0,
				"container.memory.mackerel-container-agent.usage": 1.048576e+06,
				"container.memory.mackerel-container-agent.limit": 134217728.0, // 128MiB
			},
		},
		{
			"ec2_awsvpc",
			metric.Values{
				"container.cpu.mackerel-container-agent.usage":    0.0, // Result is 0 because use the same data.
				"container.cpu.mackerel-container-agent.limit":    25.0,
				"container.cpu._internal_ecs_pause.usage":         0.0, // Result is 0 because use the same data.
				"container.cpu._internal_ecs_pause.limit":         25.0,
				"container.memory.mackerel-container-agent.usage": 1.1567104e+07,
				"container.memory.mackerel-container-agent.limit": 134217728.0, // 128MiB
				"container.memory._internal_ecs_pause.limit":      2.68435456e+08,
				"container.memory._internal_ecs_pause.usage":      573440,
			},
		},
		{
			"fargate",
			metric.Values{
				"container.cpu.mackerel-container-agent.usage":    0.0, // Result is 0 because use the same data.
				"container.cpu.mackerel-container-agent.limit":    25.0,
				"container.cpu._internal_ecs_pause.usage":         0.0, // Result is 0 because use the same data.
				"container.cpu._internal_ecs_pause.limit":         25.0,
				"container.memory.mackerel-container-agent.usage": 1.1567104e+07,
				"container.memory.mackerel-container-agent.limit": 134217728.0, // 128MiB
				"container.memory._internal_ecs_pause.limit":      2.68435456e+08,
				"container.memory._internal_ecs_pause.usage":      573440,
			},
		},
	}

	mock := &mockTaskMetadataEndpointClient{}
	ctx := context.Background()

	for _, tc := range tests {
		mock.metadataPath = fmt.Sprintf("taskmetadata/testdata/metadata_%s.json", tc.mode)
		mock.statsPath = fmt.Sprintf("taskmetadata/testdata/stats_%s.json", tc.mode)
		g := newMetricGenerator(mock, hostinfo.NewMockGenerator(3876802560.0, 8.0, nil))

		_, err := g.Generate(ctx)
		if err != nil {
			t.Errorf("Generate() should not raise error: %v", err)
		}

		got, err := g.Generate(ctx)
		if err != nil {
			t.Errorf("Generate() should not raise error: %v", err)
		}

		if !reflect.DeepEqual(tc.expected, got) {
			t.Errorf("Generate() expected %v, got %v", tc.expected, got)
		}
	}
}
