package ecsawsvpc

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	ecsTypes "github.com/aws/amazon-ecs-agent/agent/handlers/v2"

	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/platform"
	"github.com/mackerelio/mackerel-container-agent/platform/ecsawsvpc/taskmetadata"
	agentSpec "github.com/mackerelio/mackerel-container-agent/spec"
)

func TestResolvePlatform(t *testing.T) {
	tests := []struct {
		isFargate bool
		expected  string
	}{
		{true, string(platform.Fargate)},
		{false, string(platform.ECS)},
	}
	for _, tt := range tests {
		got := resolvePlatform(tt.isFargate)
		if got != tt.expected {
			t.Errorf("Platform expected %q, got %q", tt.expected, got)
		}
	}
}

func TestGenerateSpec(t *testing.T) {
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
	)
	generator := newSpecGenerator(client, true)
	got, err := generator.Generate(context.Background())
	if err != nil {
		t.Errorf("Generate() should not raise error: %v", err)
	}
	v, ok := got.(*agentSpec.CloudHostname)
	if !ok {
		t.Errorf("Generate() should return *spec.CloudHostname, got %T", got)
	}

	expected := &agentSpec.CloudHostname{
		Cloud: &mackerel.Cloud{
			Provider: string(platform.Fargate),
			MetaData: &taskSpec{
				Cluster:       "default",
				Task:          "9781c248-0edd-4cdb-9a93-f63cb662a5d3",
				TaskARN:       "arn:aws:ecs:us-west-2:012345678910:task/9781c248-0edd-4cdb-9a93-f63cb662a5d3",
				TaskFamily:    "nginx",
				TaskVersion:   "5",
				DesiredStatus: "RUNNING",
				KnownStatus:   "RUNNING",
				Containers: []container{
					{
						DockerID:   "731a0d6a3b4210e2448339bc7015aaa79bfe4fa256384f4102db86ef94cbbc4c",
						DockerName: "ecs-nginx-5-internalecspause-acc699c0cbf2d6d11700",
						Name:       "~internal~ecs~pause",
						Image:      "amazon/amazon-ecs-pause:0.1.0",
						ImageID:    "",
						Labels: map[string]string{
							"com.amazonaws.ecs.cluster":                 "default",
							"com.amazonaws.ecs.container-name":          "~internal~ecs~pause",
							"com.amazonaws.ecs.task-arn":                "arn:aws:ecs:us-west-2:012345678910:task/9781c248-0edd-4cdb-9a93-f63cb662a5d3",
							"com.amazonaws.ecs.task-definition-family":  "nginx",
							"com.amazonaws.ecs.task-definition-version": "5",
						},
						DesiredStatus: "RESOURCES_PROVISIONED",
						KnownStatus:   "RESOURCES_PROVISIONED",
						Limits: limits{
							CPU:    func(i float64) *float64 { return &i }(0),
							Memory: func(i int64) *int64 { return &i }(0),
						},
						CreatedAt: func() *time.Time { t, _ := time.Parse(time.RFC3339Nano, "2018-02-01T20:55:08.366329616Z"); return &t }(),
						StartedAt: func() *time.Time { t, _ := time.Parse(time.RFC3339Nano, "2018-02-01T20:55:09.058354915Z"); return &t }(),
						Type:      "CNI_PAUSE",
						Networks: []network{
							{
								NetworkMode:   "awsvpc",
								IPv4Addresses: []string{"10.0.2.106"},
							},
						},
					},
					{
						DockerID:   "43481a6ce4842eec8fe72fc28500c6b52edcc0917f105b83379f88cac1ff3946",
						DockerName: "ecs-nginx-5-nginx-curl-ccccb9f49db0dfe0d901",
						Name:       "nginx-curl",
						Image:      "nrdlngr/nginx-curl",
						ImageID:    "sha256:2e00ae64383cfc865ba0a2ba37f61b50a120d2d9378559dcd458dc0de47bc165",
						Labels: map[string]string{
							"com.amazonaws.ecs.cluster":                 "default",
							"com.amazonaws.ecs.container-name":          "nginx-curl",
							"com.amazonaws.ecs.task-arn":                "arn:aws:ecs:us-west-2:012345678910:task/9781c248-0edd-4cdb-9a93-f63cb662a5d3",
							"com.amazonaws.ecs.task-definition-family":  "nginx",
							"com.amazonaws.ecs.task-definition-version": "5",
						},
						DesiredStatus: "RUNNING",
						KnownStatus:   "RUNNING",
						Limits: limits{
							CPU:    func(i float64) *float64 { return &i }(0.25),
							Memory: func(i int64) *int64 { return &i }(256),
						},
						CreatedAt: func() *time.Time { t, _ := time.Parse(time.RFC3339Nano, "2018-02-01T20:55:10.554941919Z"); return &t }(),
						StartedAt: func() *time.Time { t, _ := time.Parse(time.RFC3339Nano, "2018-02-01T20:55:11.064236631Z"); return &t }(),
						Type:      "NORMAL",
						Networks: []network{
							{
								NetworkMode:   "awsvpc",
								IPv4Addresses: []string{"10.0.2.106"},
							},
						},
					},
				},
				PullStartedAt: func() *time.Time { t, _ := time.Parse(time.RFC3339Nano, "2018-02-01T20:55:09.372495529Z"); return &t }(),
				PullStoppedAt: func() *time.Time { t, _ := time.Parse(time.RFC3339Nano, "2018-02-01T20:55:10.552018345Z"); return &t }(),
				Limits: limits{
					CPU:    func(i float64) *float64 { return &i }(0.5),
					Memory: func(i int64) *int64 { return &i }(512),
				},
			},
		},
		Hostname: "9781c248-0edd-4cdb-9a93-f63cb662a5d3",
	}

	if !reflect.DeepEqual(v, expected) {
		t.Errorf("Generate() expected %v, got %v", expected, v)
	}
}
