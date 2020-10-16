package hostinfo

import (
	"testing"
)

func TestNewGenerator(t *testing.T) {
	g := NewGenerator()
	memTotal, cpuCores, err := g.Generate()
	if err != nil {
		t.Errorf("should not return error, but %v", err)
	}
	if memTotal == 0 {
		t.Error("memTotal should not be 0, but 0")
	}
	if cpuCores == 0 {
		t.Error("cpuCores should not be 0, but 0")
	}
}
