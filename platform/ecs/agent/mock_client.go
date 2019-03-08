package agent

import (
	"context"

	ecsTypes "github.com/aws/amazon-ecs-agent/agent/handlers/v1"
)

// MockClient represents a mock client of ECS introspection API
type MockClient struct {
	getInstanceMetadataCallback         func(context.Context) (*ecsTypes.MetadataResponse, error)
	getTaskMetadataWithDockerIDCallback func(context.Context, string) (*ecsTypes.TaskResponse, error)
}

// MockClientOption represents an option of mock client of ECS introspection API
type MockClientOption func(*MockClient)

// NewMockClient creates a new mock client of ECS introspection API
func NewMockClient(opts ...MockClientOption) *MockClient {
	client := &MockClient{}
	for _, opt := range opts {
		client.ApplyOption(opt)
	}
	return client
}

// ApplyOption applies a mock client option
func (c *MockClient) ApplyOption(opt MockClientOption) {
	opt(c)
}

type errCallbackNotFound string

func (err errCallbackNotFound) Error() string {
	return string(err) + " callback not found"
}

// GetInstanceMetadata ...
func (c *MockClient) GetInstanceMetadata(ctx context.Context) (*ecsTypes.MetadataResponse, error) {
	if c.getInstanceMetadataCallback != nil {
		return c.getInstanceMetadataCallback(ctx)
	}
	return nil, errCallbackNotFound("GetInstanceMetadata")
}

// MockGetInstanceMetadata returns an option to set the callback of GetInstanceMetadata
func MockGetInstanceMetadata(callback func(context.Context) (*ecsTypes.MetadataResponse, error)) MockClientOption {
	return func(c *MockClient) {
		c.getInstanceMetadataCallback = callback
	}
}

// GetTaskMetadataWithDockerID ...
func (c *MockClient) GetTaskMetadataWithDockerID(ctx context.Context, dockerID string) (*ecsTypes.TaskResponse, error) {
	if c.getTaskMetadataWithDockerIDCallback != nil {
		return c.getTaskMetadataWithDockerIDCallback(ctx, dockerID)
	}
	return nil, errCallbackNotFound("GetTaskMetadataWithDockerID")
}

// MockGetTaskMetadataWithDockerID returns an option to set the callback of GetTaskMetadataWithDockerID
func MockGetTaskMetadataWithDockerID(callback func(context.Context, string) (*ecsTypes.TaskResponse, error)) MockClientOption {
	return func(c *MockClient) {
		c.getTaskMetadataWithDockerIDCallback = callback
	}
}
