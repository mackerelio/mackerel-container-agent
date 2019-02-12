package ecsawsvpc

import (
	"context"
	"testing"

	ecsTypes "github.com/aws/amazon-ecs-agent/agent/handlers/v2"

	"github.com/mackerelio/mackerel-container-agent/platform/ecsawsvpc/taskmetadata"
)

func TestStatusRunning(t *testing.T) {
	mockClient := taskmetadata.NewMockClient()
	pform := ecsAwsvpcPlatform{client: mockClient}

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
		mockClient.ApplyOption(
			taskmetadata.MockGetMetadata(
				func(context.Context) (*ecsTypes.TaskResponse, error) {
					return &ecsTypes.TaskResponse{KnownStatus: tc.status}, nil
				},
			),
		)

		got := pform.StatusRunning(ctx)
		if got != tc.expect {
			t.Errorf("StatusRunning() expected %t, got %t", tc.expect, got)
		}
	}
}
