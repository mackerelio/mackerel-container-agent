package cgroupfs

import (
	"os"
	"path/filepath"
)

// Memory represents a part of memory subsystem
type Memory struct {
	Limit uint64
}

// Memory creates a new Memory
func (c *cgroup) Memory(subgroup string) (*Memory, error) {
	path := filepath.Join(c.root, "memory", subgroup)
	if _, err := os.Lstat(path); err != nil {
		return nil, err
	}

	limitFile, err := os.Open(filepath.Join(path, "memory.limit_in_bytes"))
	if err != nil {
		return nil, err
	}
	defer limitFile.Close()
	limit, err := readUint(limitFile)
	if err != nil {
		return nil, err
	}

	return &Memory{
		Limit: limit,
	}, nil
}
