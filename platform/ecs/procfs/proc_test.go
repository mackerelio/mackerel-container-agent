package procfs

import (
	"os"
	"testing"
)

var procfsFixtures = []interface{}{
	fsDirectory{
		Path: "30308",
		Mode: 0755,
	},
	fsSymLink{
		To:   "30308",
		Path: "self",
	},
}

func TestSelf(t *testing.T) {
	fs, _ := mockFilesystem(cgroupFixtures, "", "test_new_cgroup")
	defer os.RemoveAll(fs)

	proc, err := Self(fs)
	if err != nil {
		t.Errorf("Self() should not raise error: %v", err)
	}

	var expected = 30308
	if proc.Pid() != expected {
		t.Errorf("Pid() expected %d, got %d", proc.Pid(), expected)
	}
}

func TestNewProc(t *testing.T) {
	fs, _ := mockFilesystem(cgroupFixtures, "", "test_new_cgroup")
	defer os.RemoveAll(fs)

	var expected = 30308

	proc, err := NewProc(expected, fs)
	if err != nil {
		t.Errorf("Self() should not raise error: %v", err)
	}

	if proc.Pid() != expected {
		t.Errorf("Pid() expected %d, got %d", proc.Pid(), expected)
	}
}
