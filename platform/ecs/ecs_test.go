package ecs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/amazon-ecs-agent/agent/containermetadata"
	ecsTypes "github.com/aws/amazon-ecs-agent/agent/handlers/v2"

	"github.com/mackerelio/mackerel-container-agent/platform/ecs/internal"
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
	ctx, cancel := context.WithCancel(context.Background())
	interval := 200 * time.Millisecond

	tests := []struct {
		after            time.Duration
		callback         internal.MockTaskMetadataGetterOption
		expectCallCount  int
		expectRaiseError bool
	}{
		{
			after: 700 * time.Millisecond,
			callback: internal.MockGetTaskMetadata(
				func(ctx context.Context) (*ecsTypes.TaskResponse, error) {
					return &ecsTypes.TaskResponse{}, nil
				},
			),
			expectCallCount:  4,
			expectRaiseError: false,
		},
		{
			after: 700 * time.Millisecond,
			callback: internal.MockGetTaskMetadata(
				func(ctx context.Context) (*ecsTypes.TaskResponse, error) {
					cancel()
					return nil, errors.New("/task api already canceled")
				},
			),
			expectCallCount:  4,
			expectRaiseError: true,
		},
	}

	for _, tc := range tests {
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
			time.Sleep(tc.after)
			mock.ApplyOption(tc.callback)
		}()

		_, err := getTaskMetadata(ctx, mock, interval)

		if callCount != tc.expectCallCount {
			t.Errorf("GetTaskMetadata() expected calls %d times, but calls %d times", tc.expectCallCount, callCount)
		}

		if (err != nil) != tc.expectRaiseError {
			var not string
			if !tc.expectRaiseError {
				not = " not"
			}
			t.Errorf("GetTaskMetadata() should%s raise error: %v", not, err)
		}
	}
}
