package cgroupfs

import (
	"os"
	"reflect"
	"testing"
)

var cpuFixtures = []interface{}{
	fsDirectory{
		Path: "cpu/subgroup",
		Mode: 0755,
	},
	fsFile{
		Path:    "cpu/subgroup/cpu.cfs_quota_us",
		Content: "25000",
		Mode:    0644,
	},
	fsFile{
		Path:    "cpu/subgroup/cpu.cfs_period_us",
		Content: "100000",
		Mode:    0644,
	},
}

func TestCPU(t *testing.T) {
	fs, _ := mockFilesystem(cpuFixtures, "", "test-cgroup-cpu")
	defer os.RemoveAll(fs)

	cgroup, _ := NewCgroup(fs)
	got, err := cgroup.CPU("subgroup")
	if err != nil {
		t.Errorf("CPU() should not raise error: %v", err)
	}

	expected := &CPU{
		CfsPeriodUs: 100000,
		CfsQuotaUs:  25000,
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("CPU() expected %v, got %v", expected, got)
	}
}
