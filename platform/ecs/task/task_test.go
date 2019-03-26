package task

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"testing"

	ecsTypes "github.com/aws/amazon-ecs-agent/agent/handlers/v1"
	dockerTypes "github.com/docker/docker/api/types"
	containerTypes "github.com/docker/docker/api/types/container"

	"github.com/mackerelio/mackerel-container-agent/platform/ecs/agent"
	"github.com/mackerelio/mackerel-container-agent/platform/ecs/cgroupfs"
	"github.com/mackerelio/mackerel-container-agent/platform/ecs/docker"
	"github.com/mackerelio/mackerel-container-agent/platform/ecs/procfs"
)

var mockAgentClient = agent.NewMockClient(
	agent.MockGetTaskMetadataWithDockerID(
		func(_ context.Context, id string) (*ecsTypes.TaskResponse, error) {
			var res *ecsTypes.TaskResponse

			switch id {
			case "7e088b28bde202f19243853b0d20998a005984efa3d4b6c18e771fd149f86648":
				res = &ecsTypes.TaskResponse{
					Arn: "arn:aws:ecs:us-east-1:999999999999:task/e01d58a8-151b-40e8-bc01-22647b9ecfec",
					Containers: []ecsTypes.ContainerResponse{
						ecsTypes.ContainerResponse{
							DockerID: "7e088b28bde202f19243853b0d20998a005984efa3d4b6c18e771fd149f86648",
							Name:     "mackerel-container-agent",
						},
					},
				}

			case "docker-id":
				res = &ecsTypes.TaskResponse{
					Arn: "arn:aws:ecs:us-east-1:999999999999:task/task-id",
				}
			case "docker-id-with-new-arn":
				res = &ecsTypes.TaskResponse{
					Arn: "arn:aws:ecs:us-east-1:999999999999:task/cluster-name/task-id",
				}
			}

			if res == nil {
				return res, fmt.Errorf("invalid id: %s", id)
			}

			return res, nil
		},
	),
	agent.MockGetInstanceMetadata(
		func(context.Context) (*ecsTypes.MetadataResponse, error) {
			return &ecsTypes.MetadataResponse{
				Cluster: "mackerel-container-agent",
			}, nil
		},
	),
)

var mockCgroup = cgroupfs.NewMockCgroup(
	cgroupfs.MockCPU(
		func(subgroup string) (*cgroupfs.CPU, error) {
			var res *cgroupfs.CPU

			switch subgroup {
			case "/ecs/e01d58a8-151b-40e8-bc01-22647b9ecfec":
				res = &cgroupfs.CPU{
					CfsPeriodUs: 100000,
					CfsQuotaUs:  25000,
				}
			case "/ecs/task-id", "/ecs/cluster-name/task-id":
				res = &cgroupfs.CPU{
					CfsPeriodUs: 100000,
					CfsQuotaUs:  -1,
				}
			}

			if res == nil {
				return res, fmt.Errorf("invalid subgroup: %s", subgroup)
			}

			return res, nil
		},
	),
	cgroupfs.MockMemory(
		func(subgroup string) (*cgroupfs.Memory, error) {
			return &cgroupfs.Memory{
				Limit: 134217728,
			}, nil
		},
	),
)

var mockProc = procfs.NewMockProc(
	procfs.MockCgroup(
		func() (procfs.Cgroup, error) {
			return procfs.Cgroup{
				"memory": &procfs.CgroupLine{
					CgroupPath: "/ecs/e01d58a8-151b-40e8-bc01-22647b9ecfec/7e088b28bde202f19243853b0d20998a005984efa3d4b6c18e771fd149f86648",
				},
			}, nil
		},
	),
)

var mockDockerClient = docker.NewMockClient(
	docker.MockGetContainerStats(
		func(context.Context, string) (*dockerTypes.StatsJSON, error) {
			return &dockerTypes.StatsJSON{
				Stats: dockerTypes.Stats{
					CPUStats: dockerTypes.CPUStats{
						CPUUsage: dockerTypes.CPUUsage{
							TotalUsage: 18446744073709551615,
						},
					},
					MemoryStats: dockerTypes.MemoryStats{
						Limit: 134217728,
						Usage: 525950976,
						Stats: map[string]uint64{
							"cache": 8192,
							"rss":   1511424,
							"swap":  0,
						},
					},
				},
				Networks: map[string]dockerTypes.NetworkStats{
					"eth0": dockerTypes.NetworkStats{
						RxBytes: 25943992,
						TxBytes: 47846460,
					},
				},
			}, nil
		},
	),
	docker.MockInspectContainer(
		func(context.Context, string) (*dockerTypes.ContainerJSON, error) {
			return &dockerTypes.ContainerJSON{
				ContainerJSONBase: &dockerTypes.ContainerJSONBase{
					HostConfig: &containerTypes.HostConfig{
						Resources: containerTypes.Resources{
							CPUQuota:   0,
							CPUPercent: 0,
						},
					},
				},
			}, nil
		},
	),
)

func TestDockerID(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{
			path:     "/ecs/task-id/docker-id",
			expected: "docker-id",
		},
		{
			path:     "/ecs/cluster-name/task-id/docker-id",
			expected: "docker-id",
		},
		{
			path:     "/",
			expected: "",
		},
		{
			path:     "/ecs",
			expected: "",
		},
		{
			path:     "/ecs/task-id",
			expected: "",
		},
	}

	for _, tc := range tests {
		mockProc := procfs.NewMockProc(
			procfs.MockCgroup(
				func() (procfs.Cgroup, error) {
					return procfs.Cgroup{
						"memory": &procfs.CgroupLine{
							CgroupPath: tc.path,
						},
					}, nil
				},
			),
		)
		got, _ := getDockerID(mockProc)
		if got != tc.expected {
			t.Errorf("getDockerID() expected %v, got %v", tc.expected, got)
		}
	}
}

func TestParseTaskArn(t *testing.T) {
	tests := []struct {
		arn, expected string
	}{
		{"arn:aws:ecs:us-east-1:999999999999:task/e01d58a8-151b-40e8-bc01-22647b9ecfec", "e01d58a8-151b-40e8-bc01-22647b9ecfec"},
		{"arn:aws:ecs:::task/e01d58a8-151b-40e8-bc01-22647b9ecfec", "e01d58a8-151b-40e8-bc01-22647b9ecfec"},
		{"arn:aws:ecs:us-east-1:999999999999:e01d58a8-151b-40e8-bc01-22647b9ecfec", ""}, // no "task/" prefix
		{"invalid arn", ""},
		{"", ""},
		{"arn:aws:ecs:region:account-id:task/cluster-name/task-id", "task-id"}, // new ECS ARN. see https://aws.amazon.com/jp/blogs/compute/migrating-your-amazon-ecs-deployment-to-the-new-arn-and-resource-id-format-2/
	}

	for _, tc := range tests {
		got, _ := parseTaskArn(tc.arn)
		if got != tc.expected {
			t.Errorf("parseTaskArn() expected %v, got %v", tc.expected, got)
		}
	}
}

func TestMetadata(t *testing.T) {
	instanceMeta := &ecsTypes.MetadataResponse{
		Cluster: "mackerel-container-agent",
	}

	tests := []struct {
		proc     procfs.Proc
		expected *Metadata
	}{
		{
			mockProc,
			&Metadata{
				Arn: "arn:aws:ecs:us-east-1:999999999999:task/e01d58a8-151b-40e8-bc01-22647b9ecfec",
				Containers: []Container{
					Container{
						DockerID: "7e088b28bde202f19243853b0d20998a005984efa3d4b6c18e771fd149f86648",
						Name:     "mackerel-container-agent",
					},
				},
				Instance: instanceMeta,
				Limits: ResourceLimits{
					CPU:    25.0,
					Memory: uint64(134217728),
				},
			},
		},
		{
			procfs.NewMockProc(
				procfs.MockCgroup(
					func() (procfs.Cgroup, error) {
						return procfs.Cgroup{
							"memory": &procfs.CgroupLine{
								CgroupPath: "/ecs/task-id/docker-id",
							},
						}, nil
					},
				),
			),
			&Metadata{
				Arn:      "arn:aws:ecs:us-east-1:999999999999:task/task-id",
				Instance: instanceMeta,
				Limits: ResourceLimits{
					CPU:    float64(runtime.NumCPU()) * 100.0,
					Memory: uint64(134217728),
				},
				Containers: []Container{},
			},
		},
		{
			procfs.NewMockProc(
				procfs.MockCgroup(
					func() (procfs.Cgroup, error) {
						return procfs.Cgroup{
							"memory": &procfs.CgroupLine{
								CgroupPath: "/ecs/cluster-name/task-id/docker-id-with-new-arn",
							},
						}, nil
					},
				),
			),
			&Metadata{
				Arn:        "arn:aws:ecs:us-east-1:999999999999:task/cluster-name/task-id",
				Containers: []Container{},
				Instance:   instanceMeta,
				Limits: ResourceLimits{
					CPU:    float64(runtime.NumCPU()) * 100.0,
					Memory: uint64(134217728),
				},
			},
		},
	}

	for _, tc := range tests {
		ctx := context.Background()
		task, err := NewTaskWithProc(ctx, tc.proc, mockDockerClient, mockAgentClient, mockCgroup, nil)
		if err != nil {
			t.Errorf("newTaskWithDockerID() should not raise error: %v", err)
		}

		got, err := task.Metadata(ctx)
		if err != nil {
			t.Errorf("Metadata() should not raise error: %v", err)
		}
		if !reflect.DeepEqual(got, tc.expected) {
			t.Errorf("Metadata() expected %v, got %v", tc.expected, got)
		}
	}
}

func TestStats(t *testing.T) {
	ctx := context.Background()
	task, err := NewTaskWithProc(ctx, mockProc, mockDockerClient, mockAgentClient, mockCgroup, nil)
	if err != nil {
		t.Errorf("newTaskWithDockerID() should not raise error: %v", err)
	}

	got, err := task.Stats(ctx)
	if err != nil {
		t.Errorf("Stats() should not raise error: %v", err)
	}

	expected := map[string]*Stats{
		"mackerel-container-agent": &Stats{
			CPU: CPUStats{
				Total: 18446744073709551615,
				Limit: 0,
			},
			Memory: MemoryStats{
				Limit: 134217728,
				Usage: 525950976,
				Stats: map[string]uint64{
					"cache": 8192,
					"rss":   1511424,
					"swap":  0,
				},
			},
			Network: map[string]NetworkStats{
				"eth0": NetworkStats{
					RxBytes: 25943992,
					TxBytes: 47846460,
				},
			},
		},
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Stats() expected %v, got %v", expected, got)
	}
}

func TestIsPrivateNetworkMode(t *testing.T) {
	mockDockerClient := docker.NewMockClient()
	defer mockDockerClient.ApplyOption(
		docker.MockInspectContainer(
			func(context.Context, string) (*dockerTypes.ContainerJSON, error) {
				return nil, nil
			},
		),
	)

	tests := []struct {
		networkMode string
		expected    bool
	}{
		{"default", true},
		{"bridge", true},
		{"host", false},
		{"container:container_id", false},
		{"none", true},
	}

	for _, tc := range tests {
		ctx := context.Background()
		mockDockerClient.ApplyOption(
			docker.MockInspectContainer(
				func(context.Context, string) (*dockerTypes.ContainerJSON, error) {
					return &dockerTypes.ContainerJSON{
						ContainerJSONBase: &dockerTypes.ContainerJSONBase{
							HostConfig: &containerTypes.HostConfig{
								NetworkMode: containerTypes.NetworkMode(tc.networkMode),
							},
						},
					}, nil
				},
			),
		)
		tsk, _ := NewTaskWithProc(ctx, mockProc, mockDockerClient, mockAgentClient, mockCgroup, nil)

		got := tsk.IsPrivateNetworkMode()
		if got != tc.expected {
			t.Errorf("NetworkMode() expected %v, got %v", tc.expected, got)
		}
	}
}

func TestIgnoreContainer(t *testing.T) {
	mockAgentClient := agent.NewMockClient(
		agent.MockGetTaskMetadataWithDockerID(
			func(context.Context, string) (*ecsTypes.TaskResponse, error) {
				return &ecsTypes.TaskResponse{
					Arn: "arn:aws:ecs:us-east-1:999999999999:task/dummy",
					Containers: []ecsTypes.ContainerResponse{
						ecsTypes.ContainerResponse{Name: "foo"},
						ecsTypes.ContainerResponse{Name: "bar"},
						ecsTypes.ContainerResponse{Name: "baz"},
					},
				}, nil
			},
		),
		agent.MockGetInstanceMetadata(
			func(context.Context) (*ecsTypes.MetadataResponse, error) {
				return &ecsTypes.MetadataResponse{
					Cluster: "mackerel-container-agent",
				}, nil
			},
		),
	)

	tests := []struct {
		ignoreContainer *regexp.Regexp
		expected        int
	}{
		{nil, 3},
		{regexp.MustCompile(`\Afoo\z`), 2},
		{regexp.MustCompile(`a`), 1},
		{regexp.MustCompile(``), 0},
	}

	for _, tc := range tests {
		ctx := context.Background()
		task, err := NewTaskWithProc(ctx, mockProc, mockDockerClient, mockAgentClient, mockCgroup, tc.ignoreContainer)
		if err != nil {
			t.Errorf("newTaskWithDockerID() should not raise error: %v", err)
		}

		meta, err := task.Metadata(ctx)
		got := len(meta.Containers)
		if got != tc.expected {
			t.Errorf("meta.Containers expected %d containers, got %v containers", tc.expected, got)
		}

		stats, err := task.Stats(ctx)
		got = len(stats)
		if got != tc.expected {
			t.Errorf("Stats() expected %d containers, got %v containers", tc.expected, got)
		}
	}

}
