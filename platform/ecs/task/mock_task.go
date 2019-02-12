package task

import "context"

// MockTaskOption is functional option.
type MockTaskOption func(*MockTask)

// MockTask is mock cgroup.
type MockTask struct {
	metadataCallback             func(context.Context) (*Metadata, error)
	statsCallback                func(context.Context) (map[string]*Stats, error)
	isPrivateNetworkModeCallback func() bool
}

// NewMockTask creates a new mock Task.
func NewMockTask(opts ...MockTaskOption) *MockTask {
	proc := &MockTask{}
	for _, opt := range opts {
		proc.ApplyOption(opt)
	}
	return proc
}

// Metadata returns task metadata.
func (p *MockTask) Metadata(ctx context.Context) (*Metadata, error) {
	if p.metadataCallback != nil {
		return p.metadataCallback(ctx)
	}
	return nil, nil
}

// Stats returns task stats.
func (p *MockTask) Stats(ctx context.Context) (map[string]*Stats, error) {
	if p.statsCallback != nil {
		return p.statsCallback(ctx)
	}
	return nil, nil
}

// IsPrivateNetworkMode returns true when task is private network mode.
func (p *MockTask) IsPrivateNetworkMode() bool {
	if p.isPrivateNetworkModeCallback != nil {
		return p.isPrivateNetworkModeCallback()
	}
	return false
}

// MockMetadata returns MockTaskOption to replace Metadata().
func MockMetadata(callback func(context.Context) (*Metadata, error)) MockTaskOption {
	return func(c *MockTask) {
		c.metadataCallback = callback
	}
}

// MockStats returns MockTaskOption to replace Stats().
func MockStats(callback func(context.Context) (map[string]*Stats, error)) MockTaskOption {
	return func(c *MockTask) {
		c.statsCallback = callback
	}
}

// MockIsPrivateNetworkMode returns MockTaskOption to replace IsPrivateNetworkMode().
func MockIsPrivateNetworkMode(callback func() bool) MockTaskOption {
	return func(c *MockTask) {
		c.isPrivateNetworkModeCallback = callback
	}
}

// ApplyOption apply MockTaskOption.
func (p *MockTask) ApplyOption(opt MockTaskOption) {
	opt(p)
}
