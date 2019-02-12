package cgroupfs

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestNewCgroup(t *testing.T) {
	root, _ := ioutil.TempDir("", "test-new-cgroup")
	defer os.Remove(root)

	var err error

	_, err = NewCgroup(root)
	if err != nil {
		t.Errorf("NewCgroup() should not raise error: %v", err)
	}

	_, err = NewCgroup("")
	if err == nil {
		t.Errorf("NewCgroup() should raise error: %v", err)
	}
}
