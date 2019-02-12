package procfs

// MockProcOption is functional option.
type MockProcOption func(*MockProc)

// MockProc is mock cgroup.
type MockProc struct {
	pidCallback       func() int
	cgroupCallback func() (Cgroup, error)
}

// NewMockProc creates a new mock Proc.
func NewMockProc(opts ...MockProcOption) *MockProc {
	proc := &MockProc{}
	for _, opt := range opts {
		proc.ApplyOption(opt)
	}
	return proc
}

// Pid returns process id.
func (p *MockProc) Pid() int {
	if p.pidCallback != nil {
		return p.pidCallback()
	}
	return 0
}

// Cgroup returns Cgroup.
func (p *MockProc) Cgroup() (Cgroup, error) {
	if p.cgroupCallback != nil {
		return p.cgroupCallback()
	}
	return nil, nil
}

// MockPid returns MockProcOption to replace Pid().
func MockPid(callback func() int) MockProcOption {
	return func(c *MockProc) {
		c.pidCallback = callback
	}
}

// MockCgroup returns MockProcOption to replace Cgroup().
func MockCgroup(callback func() (Cgroup, error)) MockProcOption {
	return func(c *MockProc) {
		c.cgroupCallback = callback
	}
}

// ApplyOption apply MockProcOption.
func (p *MockProc) ApplyOption(opt MockProcOption) {
	opt(p)
}
