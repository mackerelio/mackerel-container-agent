package internal

import (
	"context"
	"errors"

	ecsTypes "github.com/aws/amazon-ecs-agent/agent/handlers/v2"
)

// MockTaskMetadataGetter is a mock of /task API endpoint
type MockTaskMetadataGetter struct {
	getTaskMetadataCallback func(context.Context) (*ecsTypes.TaskResponse, error)
}

// MockTaskMetadataGetterOption represents an option of mock client of /task API endpoint
type MockTaskMetadataGetterOption func(*MockTaskMetadataGetter)

// NewMockTaskMetadataGetter creates a new mock of /task API endpoint
func NewMockTaskMetadataGetter(opts ...MockTaskMetadataGetterOption) *MockTaskMetadataGetter {
	g := &MockTaskMetadataGetter{}
	for _, o := range opts {
		g.ApplyOption(o)
	}
	return g
}

// ApplyOption applies a mock option
func (g *MockTaskMetadataGetter) ApplyOption(opt MockTaskMetadataGetterOption) {
	opt(g)
}

// GetTaskMetadata returns /task API response
func (g *MockTaskMetadataGetter) GetTaskMetadata(ctx context.Context) (*ecsTypes.TaskResponse, error) {
	if g.getTaskMetadataCallback != nil {
		return g.getTaskMetadataCallback(ctx)
	}
	return nil, errors.New("MockGetTaskMetadata not found")
}

// MockGetTaskMetadata returns an option to set the callback of GetTaskMetadata
func MockGetTaskMetadata(callback func(context.Context) (*ecsTypes.TaskResponse, error)) MockTaskMetadataGetterOption {
	return func(g *MockTaskMetadataGetter) {
		g.getTaskMetadataCallback = callback
	}
}
