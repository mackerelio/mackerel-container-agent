package ecsv3

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	ecsTypes "github.com/aws/amazon-ecs-agent/agent/handlers/v2"
	"github.com/mackerelio/mackerel-container-agent/platform"
	agentSpec "github.com/mackerelio/mackerel-container-agent/spec"
)

type mockTaskMetadataFetcher struct {
	path string
}

func (m *mockTaskMetadataFetcher) FetchTaskMetadata(ctx context.Context) (*ecsTypes.TaskResponse, error) {
	f, err := os.Open(m.path)
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

func TestGenerate(t *testing.T) {
	tests := []struct {
		path      string
		isFargate bool
		expected  platform.Type
	}{
		{"testdata/metadata_ec2_bridge.json", false, platform.ECS},
		{"testdata/metadata_ec2_host.json", false, platform.ECS},
		{"testdata/metadata_ec2_awsvpc.json", false, platform.ECS},
		{"testdata/metadata_fargate.json", true, platform.Fargate},
	}

	mock := &mockTaskMetadataFetcher{}
	ctx := context.Background()

	for _, tc := range tests {
		mock.path = tc.path
		g := newSpecGenerator(mock, tc.isFargate)

		spec, err := g.Generate(ctx)
		if err != nil {
			t.Errorf("Generate() should not raise error: %v", err)
		}

		got, ok := spec.(*agentSpec.CloudHostname)
		if !ok {
			t.Errorf("Generate() should return *spec.CloudHostname, got %T", got)
		}

		if got.Hostname != "task-id" {
			t.Errorf("Hostname expected %v, got %v", "task-id", got)
		}
		if got.Cloud.Provider != string(tc.expected) {
			t.Errorf("Provider expected %v, got %v", tc.expected, got.Cloud.Provider)
		}
		if got.Cloud.MetaData == nil {
			t.Error("MetaData should not be nil")
		}
		t.Logf("%+v\n\n", got.Cloud.MetaData)
	}
}

func TestResolvePlatform(t *testing.T) {
	tests := []struct {
		isFargate bool
		expected  platform.Type
	}{
		{true, platform.Fargate},
		{false, platform.ECS},
	}

	for _, tc := range tests {
		got := resolvePlatform(tc.isFargate)
		if got != string(tc.expected) {
			t.Errorf("resolvePlatform() expected %v, got %v", tc.expected, got)
		}
	}
}

func TestGetTaskID(t *testing.T) {
	tests := []struct {
		taskARN  string
		expected string
	}{
		{"arn:aws:ecs:us-east-1:012345678910:task/task-id", "task-id"},
		{"arn:aws:ecs:us-east-1:012345678910:task/cluster-name/task-id", "task-id"},
	}

	for _, tc := range tests {
		got, err := getTaskID(tc.taskARN)
		if err != nil {
			t.Errorf("getTaskID() should not raise error: %v", err)
		}
		if got != tc.expected {
			t.Errorf("getTaskID() expected %v, got %v", tc.expected, got)
		}
	}

}
