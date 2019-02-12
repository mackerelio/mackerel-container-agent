package procfs

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// CgroupLine represents a line in /proc/PID/cgroup
type CgroupLine struct {
	HierarchyID    int
	ControllerList string
	CgroupPath     string
}

// Cgroup represents /proc/PID/cgroup
type Cgroup map[string]*CgroupLine

func (Cgroup) parseLine(line string) (*CgroupLine, error) {
	parts := strings.Split(line, ":")
	if len(parts) != 3 {
		return nil, errors.New("invalid cgroup line")
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, err
	}

	return &CgroupLine{
		HierarchyID:    id,
		ControllerList: parts[1],
		CgroupPath:     parts[2],
	}, nil
}

// Cgroup creates a new Cgroup
func (p *proc) Cgroup() (Cgroup, error) {
	return cgroup(filepath.Join(p.path, "cgroup"))
}

func cgroup(path string) (Cgroup, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cgroup = make(Cgroup, 0)
	s := bufio.NewScanner(file)
	for s.Scan() {
		line, err := cgroup.parseLine(s.Text())
		if err != nil {
			return nil, err
		}
		cgroup[line.ControllerList] = line
	}

	return cgroup, s.Err()
}
