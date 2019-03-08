package kubelet

import "context"

// MockClient represents a mock client of Kubelet APIs
type MockClient struct {
	getPodCallback      func(context.Context) (*Pod, error)
	getPodStatsCallback func(context.Context) (*PodStats, error)
	getSpecCallback     func(context.Context) (*MachineInfo, error)
}

// MockClientOption represents an option of mock client of Kubelet APIs
type MockClientOption func(*MockClient)

// NewMockClient creates a new mock client of Kubelet APIs
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

type errCallbackNotFound string

func (err errCallbackNotFound) Error() string {
	return string(err) + " callback not found"
}

// GetPod ...
func (c *MockClient) GetPod(ctx context.Context) (*Pod, error) {
	if c.getPodCallback != nil {
		return c.getPodCallback(ctx)
	}
	return nil, errCallbackNotFound("GetPod")
}

// MockGetPod returns an option to set the callback of GetPod
func MockGetPod(callback func(context.Context) (*Pod, error)) MockClientOption {
	return func(c *MockClient) {
		c.getPodCallback = callback
	}
}

// GetPodStats ...
func (c *MockClient) GetPodStats(ctx context.Context) (*PodStats, error) {
	if c.getPodStatsCallback != nil {
		return c.getPodStatsCallback(ctx)
	}
	return nil, errCallbackNotFound("GetPodStats")
}

// MockGetPodStats returns an option to set the callback of GetPodStats
func MockGetPodStats(callback func(context.Context) (*PodStats, error)) MockClientOption {
	return func(c *MockClient) {
		c.getPodStatsCallback = callback
	}
}

// GetSpec ...
func (c *MockClient) GetSpec(ctx context.Context) (*MachineInfo, error) {
	if c.getSpecCallback != nil {
		return c.getSpecCallback(ctx)
	}
	return nil, errCallbackNotFound("GetSpec")
}

// MockGetSpec returns an option to set the callback of GetSpec
func MockGetSpec(callback func(context.Context) (*MachineInfo, error)) MockClientOption {
	return func(c *MockClient) {
		c.getSpecCallback = callback
	}
}
