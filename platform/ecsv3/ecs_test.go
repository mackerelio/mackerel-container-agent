package ecsv3

import (
	"testing"

	"github.com/aws/amazon-ecs-agent/agent/containermetadata"
	ecsTypes "github.com/aws/amazon-ecs-agent/agent/handlers/v2"
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
