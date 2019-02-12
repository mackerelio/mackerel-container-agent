package taskmetadata

import (
	"context"

	ecsTypes "github.com/aws/amazon-ecs-agent/agent/handlers/v2"
	dockerTypes "github.com/docker/docker/api/types"
)

// MockClient a mock client of ECS task metadata endpoint
type MockClient struct {
	getMetadataCallback func(context.Context) (*ecsTypes.TaskResponse, error)
	getStatsCallback    func(context.Context) (map[string]*dockerTypes.Stats, error)
}

// MockClientOption represents an option of mock client of ECS task metadata endpoint
type MockClientOption func(*MockClient)

// NewMockClient creates a new mock client of ECS task metadata endpoint
func NewMockClient(opts ...MockClientOption) *MockClient {
	c := &MockClient{}
	for _, o := range opts {
		c.ApplyOption(o)
	}
	return c
}

// ApplyOption applies a mock client option
func (c *MockClient) ApplyOption(opt MockClientOption) {
	opt(c)
}

// GetMetadata ...
func (c *MockClient) GetMetadata(ctx context.Context) (*ecsTypes.TaskResponse, error) {
	if c.getMetadataCallback != nil {
		return c.getMetadataCallback(ctx)
	}
	return nil, nil
}

// MockGetMetadata returns an option to set the callback of GetMetadata
func MockGetMetadata(callback func(context.Context) (*ecsTypes.TaskResponse, error)) MockClientOption {
	return func(c *MockClient) {
		c.getMetadataCallback = callback
	}
}

// GetStats ...
func (c *MockClient) GetStats(ctx context.Context) (map[string]*dockerTypes.Stats, error) {
	if c.getStatsCallback != nil {
		return c.getStatsCallback(ctx)
	}
	return nil, nil
}

// MockGetStats returns an option to set the callback of GetStats
func MockGetStats(callback func(context.Context) (map[string]*dockerTypes.Stats, error)) MockClientOption {
	return func(c *MockClient) {
		c.getStatsCallback = callback
	}
}
