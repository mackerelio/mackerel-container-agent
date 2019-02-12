package cgroupfs

import (
	"os"
	"path/filepath"
)

// Cgroup interface gets subsystems
type Cgroup interface {
	CPU(string) (*CPU, error)
	Memory(string) (*Memory, error)
}

type cgroup struct {
	root string
}

// NewCgroup creates a new Cgroup
func NewCgroup(root string) (Cgroup, error) {
	if _, err := os.Lstat(filepath.Join(root)); err != nil {
		return nil, err
	}
	return &cgroup{root: root}, nil
}
