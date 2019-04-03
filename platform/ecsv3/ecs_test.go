package ecsv3

import (
	"testing"

	"github.com/mackerelio/mackerel-container-agent/provider"
)

func TestIsRunning(t *testing.T) {
	tests := []struct {
		status string
		expect bool
	}{
		{"running", true},
		{"Running", true},
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
		expect       provider.Type
	}{
		{"AWS_ECS_FARGATE", provider.Fargate},
		{"AWS_ECS_EC2", provider.ECS},
		{"unknown", provider.Type("UNKNOWN")},
		{"", provider.Type("UNKNOWN")},
	}

	for _, tc := range tests {
		got, _ := resolveProvider(tc.executionEnv)
		if got != tc.expect {
			t.Errorf("resolveProvider() expected %v, got %v", tc.expect, got)
		}
	}

}
