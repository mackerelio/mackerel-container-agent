package ecsv3

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/amazon-ecs-agent/agent/containermetadata"
	ecsTypes "github.com/aws/amazon-ecs-agent/agent/handlers/v2"

	"github.com/mackerelio/mackerel-container-agent/platform/ecsv3/internal"
)

func TestIsRunning(t *testing.T) {
	tests := []struct {
		status string
		expect bool
	}{
		{"running", false},
		{"Running", false},
		{"RUNNING", true},
		{"PENDING", false},
		{"", false},
	}

	for _, tc := range tests {
		got := isRunning(tc.status)
		if got != tc.expect {
			t.Errorf("isRunning() expected %t, got %t", tc.expect, got)
		}
	}
}

func TestResolveProvider(t *testing.T) {
	tests := []struct {
		executionEnv string
		expect       provider
	}{
		{"AWS_ECS_FARGATE", fargateProvider},
		{"AWS_ECS_EC2", ecsProvider},
		{"unknown", provider("UNKNOWN")},
		{"", provider("UNKNOWN")},
	}

	for _, tc := range tests {
		got, _ := resolveProvider(tc.executionEnv)
		if got != tc.expect {
			t.Errorf("resolveProvider() expected %v, got %v", tc.expect, got)
		}
	}

}

func TestDetectNetworkMode(t *testing.T) {
	type expect struct {
		networkMode networkMode
		raiseError  bool
	}

	tests := []struct {
		meta   *ecsTypes.TaskResponse
		expect expect
	}{
		{
			&ecsTypes.TaskResponse{},
			expect{"", true},
		},
		{
			&ecsTypes.TaskResponse{
				Containers: []ecsTypes.ContainerResponse{
					ecsTypes.ContainerResponse{},
				},
			},
			expect{"", true},
		},
		{
			&ecsTypes.TaskResponse{
				Containers: []ecsTypes.ContainerResponse{
					ecsTypes.ContainerResponse{
						Networks: []containermetadata.Network{
							containermetadata.Network{},
						},
					},
				},
			},
			expect{"", true},
		},
		{
			&ecsTypes.TaskResponse{
				Containers: []ecsTypes.ContainerResponse{
					ecsTypes.ContainerResponse{
						Networks: []containermetadata.Network{
							containermetadata.Network{NetworkMode: "default"},
						},
					},
				},
			},
			expect{bridgeNetworkMode, false},
		},
		{
			&ecsTypes.TaskResponse{
				Containers: []ecsTypes.ContainerResponse{
					ecsTypes.ContainerResponse{
						Networks: []containermetadata.Network{
							containermetadata.Network{NetworkMode: "bridge"},
						},
					},
				},
			},
			expect{bridgeNetworkMode, false},
		},
		{
			&ecsTypes.TaskResponse{
				Containers: []ecsTypes.ContainerResponse{
					ecsTypes.ContainerResponse{
						Networks: []containermetadata.Network{
							containermetadata.Network{NetworkMode: "host"},
						},
					},
				},
			},
			expect{hostNetworkMode, false},
		},
		{
			&ecsTypes.TaskResponse{
				Containers: []ecsTypes.ContainerResponse{
					ecsTypes.ContainerResponse{
						Networks: []containermetadata.Network{
							containermetadata.Network{NetworkMode: "awsvpc"},
						},
					},
				},
			},
			expect{awsvpcNetworkMode, false},
		},
	}

	for _, tc := range tests {
		got, err := detectNetworkMode(tc.meta)

		if got != tc.expect.networkMode {
			t.Errorf("detectNetworkMode() expected %v, got %v", tc.expect.networkMode, got)
		}

		if (err != nil) != tc.expect.raiseError {
			var msg string
			if !tc.expect.raiseError {
				msg = "not "
			}

			t.Errorf("detectNetworkMode() should %sraise error: %v", msg, err)
		}
	}

}

func TestGetTaskMetadata(t *testing.T) {
	ctx := context.Background()
	interval := 200 * time.Millisecond

	var callCount int
	mock := internal.NewMockTaskMetadataGetter(
		internal.MockGetTaskMetadata(
			func(ctx context.Context) (*ecsTypes.TaskResponse, error) {
				callCount++
				return nil, errors.New("/task api error")
			},
		),
	)

	go func() {
		time.Sleep(700 * time.Millisecond)
		mock.ApplyOption(
			internal.MockGetTaskMetadata(
				func(ctx context.Context) (*ecsTypes.TaskResponse, error) {
					return &ecsTypes.TaskResponse{}, nil
				},
			),
		)
	}()

	getTaskMetadata(ctx, mock, interval)

	if expected := 4; callCount != expected {
		t.Errorf("GetTaskMetadata() expected calls %d times, but calls %d times", expected, callCount)
	}
}
