package cgroupfs

import (
	"os"
	"path/filepath"
)

// CPU represents a part of cpu subsystem
type CPU struct {
	CfsQuotaUs, CfsPeriodUs int64
}

// CPU creates a new CPU
func (c *cgroup) CPU(subgroup string) (*CPU, error) {
	path := filepath.Join(c.root, "cpu", subgroup)
	if _, err := os.Lstat(path); err != nil {
		return nil, err
	}

	quotaFile, err := os.Open(filepath.Join(path, "cpu.cfs_quota_us"))
	if err != nil {
		return nil, err
	}
	defer quotaFile.Close()
	quota, err := readInt(quotaFile)
	if err != nil {
		return nil, err
	}

	periodFile, err := os.Open(filepath.Join(path, "cpu.cfs_period_us"))
	if err != nil {
		return nil, err
	}
	defer periodFile.Close()
	period, err := readInt(periodFile)
	if err != nil {
		return nil, err
	}

	return &CPU{
		CfsQuotaUs:  quota,
		CfsPeriodUs: period,
	}, nil
}
