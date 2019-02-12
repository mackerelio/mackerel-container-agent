package ecs

import (
	"context"
	"reflect"
	"runtime"
	"testing"

	ecsTypes "github.com/aws/amazon-ecs-agent/agent/handlers/v1"

	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/platform"
	"github.com/mackerelio/mackerel-container-agent/platform/ecs/task"
	agentSpec "github.com/mackerelio/mackerel-container-agent/spec"
)

func TestGenerateSpec(t *testing.T) {
	mockTask := task.NewMockTask(
		task.MockMetadata(
			func(context.Context) (*task.Metadata, error) {
				return &task.Metadata{
					Arn: "arn:aws:ecs:us-east-1:999999999999:task/e01d58a8-151b-40e8-bc01-22647b9ecfec",
					Containers: []task.Container{
						{
							DockerID:   "79c796ed2a7f864f485c76f83f3165488097279d296a7c05bd5201a1c69b2920",
							DockerName: "ecs-nginx-efs-2-nginx-9ac0808dd0afa495f001",
							Name:       "nginx",
						},
						{
							DockerID:   "7e088b28bde202f19243853b0d20998a005984efa3d4b6c18e771fd149f86648",
							DockerName: "ecs-mackerel-container-agent-7-mackerel-container-agent-96b2f7c0c7c2ccff9101",
							Name:       "mackerel-container-agent",
						},
					},
					DesiredStatus: "RUNNING",
					Family:        "nginx-develop",
					KnownStatus:   "RUNNING",
					Version:       "2",
					Instance: &ecsTypes.MetadataResponse{
						Cluster:              "mackerel-container-agent",
						ContainerInstanceArn: func(s string) *string { return &s }("arn:aws:ecs:ap-northeast-1:999999999999:container-instance/07ed8509-6b38-4b36-b252-d9fb856c2a83"),
						Version:              "Amazon ECS Agent - v1.18.0 (c0defea9)",
					},
					Limits: task.ResourceLimits{
						CPU:    float64(runtime.NumCPU()),
						Memory: uint64(134217728),
					},
				}, nil
			},
		),
	)

	generator := newSpecGenerator(mockTask)

	got, err := generator.Generate(context.Background())
	if err != nil {
		t.Errorf("Generate() should not raise error: %v", err)
	}

	v, ok := got.(*agentSpec.CloudHostname)
	if !ok {
		t.Errorf("Generate() should return *spec.CloudHostname got %T", got)
	}
	if v.Cloud.Provider != "ecs" {
		t.Errorf("Provider should %q, got %q", "ecs", v.Cloud.Provider)
	}

	expected := &agentSpec.CloudHostname{
		Cloud: &mackerel.Cloud{
			Provider: string(platform.ECS),
			MetaData: &ecsSpec{
				Cluster:       "mackerel-container-agent",
				Task:          "e01d58a8-151b-40e8-bc01-22647b9ecfec",
				TaskARN:       "arn:aws:ecs:us-east-1:999999999999:task/e01d58a8-151b-40e8-bc01-22647b9ecfec",
				TaskFamily:    "nginx-develop",
				TaskVersion:   "2",
				DesiredStatus: "RUNNING",
				KnownStatus:   "RUNNING",
				Containers: []container{
					container{
						DockerID:   "79c796ed2a7f864f485c76f83f3165488097279d296a7c05bd5201a1c69b2920",
						DockerName: "ecs-nginx-efs-2-nginx-9ac0808dd0afa495f001",
						Name:       "nginx",
					},
					container{
						DockerID:   "7e088b28bde202f19243853b0d20998a005984efa3d4b6c18e771fd149f86648",
						DockerName: "ecs-mackerel-container-agent-7-mackerel-container-agent-96b2f7c0c7c2ccff9101",
						Name:       "mackerel-container-agent",
					},
				},
				Limits: resourceLimits{
					CPU:    float64(runtime.NumCPU()),
					Memory: uint64(134217728),
				},
			},
		},
		Hostname: "e01d58a8-151b-40e8-bc01-22647b9ecfec",
	}

	if !reflect.DeepEqual(v, expected) {
		t.Errorf("Generate() expected %v, got %v", expected, v)
	}
}
