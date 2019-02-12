package cgroupfs

import (
	"os"
	"reflect"
	"testing"
)

var memoryFixtures = []interface{}{
	fsDirectory{
		Path: "memory/subgroup",
		Mode: 0755,
	},
	fsFile{
		Path:    "memory/subgroup/memory.limit_in_bytes",
		Content: "134217728",
		Mode:    0644,
	},
}

func TestMemory(t *testing.T) {
	fs, _ := mockFilesystem(memoryFixtures, "", "test-cgroup-memory")
	defer os.RemoveAll(fs)

	cgroup, _ := NewCgroup(fs)
	got, err := cgroup.Memory("subgroup")
	if err != nil {
		t.Errorf("memory() should not raise error: %v", err)
	}

	expected := &Memory{
		Limit: 134217728,
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("memory() expected %v, got %v", expected, got)
	}
}
