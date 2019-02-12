package cgroupfs

// MockCgroupOption is functional option
type MockCgroupOption func(*MockCgroup)

// MockCgroup is mock cgroup
type MockCgroup struct {
	cpuCallback    func(string) (*CPU, error)
	memoryCallback func(string) (*Memory, error)
}

// NewMockCgroup creates a new mock Cgroup
func NewMockCgroup(opts ...MockCgroupOption) *MockCgroup {
	client := &MockCgroup{}
	for _, opt := range opts {
		client.ApplyOption(opt)
	}
	return client
}

// CPU returns cpu resource
func (c *MockCgroup) CPU(subgroup string) (*CPU, error) {
	if c.cpuCallback != nil {
		return c.cpuCallback(subgroup)
	}
	return nil, nil
}

// Memory returns cpu resource
func (c *MockCgroup) Memory(subgroup string) (*Memory, error) {
	if c.memoryCallback != nil {
		return c.memoryCallback(subgroup)
	}
	return nil, nil
}

// MockCPU returns MockCgroupOption for cpuCallback
func MockCPU(callback func(string) (*CPU, error)) MockCgroupOption {
	return func(c *MockCgroup) {
		c.cpuCallback = callback
	}
}

// MockMemory returns MockCgroupOption for cpuCallback
func MockMemory(callback func(string) (*Memory, error)) MockCgroupOption {
	return func(c *MockCgroup) {
		c.memoryCallback = callback
	}
}

// ApplyOption apply MockCgroupOption
func (c *MockCgroup) ApplyOption(opt MockCgroupOption) {
	opt(c)
}
