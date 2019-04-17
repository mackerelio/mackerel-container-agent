package ecsv3

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	ecsTypes "github.com/aws/amazon-ecs-agent/agent/handlers/v2"
	dockerTypes "github.com/docker/docker/api/types"
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
		mode string
	}{
		{"ec2_bridge"},
		{"ec2_host"},
		{"ec2_awsvpc"},
		{"fargate"},
	}

	mock := &mockTaskMetadataEndpointClient{}
	ctx := context.Background()

	for _, tc := range tests {
		mock.metadataPath = fmt.Sprintf("taskmetadata/testdata/metadata_%s.json", tc.mode)
		mock.statsPath = fmt.Sprintf("taskmetadata/testdata/stats_%s.json", tc.mode)
		g := newMetricGenerator(mock)

		_, err := g.Generate(ctx)
		if err != nil {
			t.Errorf("Generate() should not raise error: %v", err)
		}

		got, err := g.Generate(ctx)
		if err != nil {
			t.Errorf("Generate() should not raise error: %v", err)
		}

		t.Logf("%s = %#v\n", tc.mode, got)
	}
}
