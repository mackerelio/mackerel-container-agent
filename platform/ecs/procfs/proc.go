package procfs

import (
	"os"
	"path/filepath"
	"strconv"
)

const (
	// DefaultMountPoint represents procfs munt point
	DefaultMountPoint = "/proc"
)

// Proc interface gets cgroup
type Proc interface {
	Pid() int
	Cgroup() (Cgroup, error)
}

type proc struct {
	pid  int
	path string
}

// Self creates a new Proc represents the current process
func Self(mountPoint string) (Proc, error) {
	if mountPoint == "" {
		mountPoint = DefaultMountPoint
	}
	path := filepath.Join(mountPoint, "self")
	p, err := os.Readlink(path)
	if err != nil {
		return nil, err
	}
	pid, err := strconv.Atoi(filepath.Base(p))
	if err != nil {
		return nil, err
	}
	return NewProc(pid, mountPoint)
}

// NewProc creates a new Proc
func NewProc(pid int, mountPoint string) (Proc, error) {
	if mountPoint == "" {
		mountPoint = DefaultMountPoint
	}
	path := filepath.Join(mountPoint, strconv.Itoa(pid))
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	return &proc{
		pid:  pid,
		path: path,
	}, nil
}

// Pid returns process id.
func (p *proc) Pid() int {
	return p.pid
}
