package docker

import (
	"context"

	dockerTypes "github.com/docker/docker/api/types"
)

// MockClient represents a mock client of ECS container stats and specs
type MockClient struct {
	getContainerStatsCallback func(context.Context, string) (*dockerTypes.StatsJSON, error)
	inspectContainerCallback  func(context.Context, string) (*dockerTypes.ContainerJSON, error)
}

// MockClientOption represents an option of mock client of ECS container stats and specs
type MockClientOption func(*MockClient)

// NewMockClient creates a new mock client of ECS container stats and specs
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

// GetContainerStats ...
func (c *MockClient) GetContainerStats(ctx context.Context, id string) (*dockerTypes.StatsJSON, error) {
	if c.getContainerStatsCallback != nil {
		return c.getContainerStatsCallback(ctx, id)
	}
	return nil, errCallbackNotFound("GetContainerStats")
}

// MockGetContainerStats returns an option to set the callback of GetContainerStats
func MockGetContainerStats(callback func(context.Context, string) (*dockerTypes.StatsJSON, error)) MockClientOption {
	return func(c *MockClient) {
		c.getContainerStatsCallback = callback
	}
}

// InspectContainer ...
func (c *MockClient) InspectContainer(ctx context.Context, id string) (*dockerTypes.ContainerJSON, error) {
	if c.inspectContainerCallback != nil {
		return c.inspectContainerCallback(ctx, id)
	}
	return nil, errCallbackNotFound("InspectContainer")
}

// MockInspectContainer returns an option to set the callback of InspectContainer
func MockInspectContainer(callback func(context.Context, string) (*dockerTypes.ContainerJSON, error)) MockClientOption {
	return func(c *MockClient) {
		c.inspectContainerCallback = callback
	}
}
