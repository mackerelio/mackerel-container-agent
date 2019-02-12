package ecs

import (
	"context"
	"testing"

	"github.com/mackerelio/mackerel-container-agent/platform/ecs/task"
)

func TestStatusRunning(t *testing.T) {
	mockTask := task.NewMockTask()
	pform := ecsPlatform{mockTask}

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
		ctx := context.Background()
		mockTask.ApplyOption(
			task.MockMetadata(
				func(context.Context) (*task.Metadata, error) {
					return &task.Metadata{KnownStatus: tc.status}, nil
				},
			),
		)

		got := pform.StatusRunning(ctx)
		if got != tc.expect {
			t.Errorf("StatusRunning() expected %t, got %t", tc.expect, got)
		}
	}
}
