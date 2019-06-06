package ecs

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	ecsTypes "github.com/aws/amazon-ecs-agent/agent/handlers/v2"
	"github.com/mackerelio/mackerel-container-agent/platform/ecs/internal"
	agentSpec "github.com/mackerelio/mackerel-container-agent/spec"
)

func TestGenerateSpec(t *testing.T) {
	tests := []struct {
		path     string
		provider provider
	}{
		{"taskmetadata/testdata/metadata_ec2_bridge.json", ecsProvider},
		{"taskmetadata/testdata/metadata_ec2_host.json", ecsProvider},
		{"taskmetadata/testdata/metadata_ec2_awsvpc.json", ecsProvider},
		{"taskmetadata/testdata/metadata_fargate.json", fargateProvider},
	}

	var path string
	mock := internal.NewMockTaskMetadataGetter(
		internal.MockGetTaskMetadata(
			func(ctx context.Context) (*ecsTypes.TaskResponse, error) {
				f, err := os.Open(path)
				if err != nil {
					return nil, err
				}
				defer f.Close()
				var res ecsTypes.TaskResponse
				if err := json.NewDecoder(f).Decode(&res); err != nil {
					return nil, err
				}
				return &res, nil
			},
		),
	)

	ctx := context.Background()

	for _, tc := range tests {
		path = tc.path

		g := newSpecGenerator(mock, tc.provider)

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
		if got.Cloud.MetaData == nil {
			t.Error("MetaData should not be nil")
		}
		t.Logf("%+v\n\n", got.Cloud.MetaData)
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
