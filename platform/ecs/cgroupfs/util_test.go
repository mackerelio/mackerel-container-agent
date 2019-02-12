package cgroupfs

import (
	"strings"
	"testing"
)

func TestReadInt(t *testing.T) {
	tests := []struct {
		data     string
		expected int64
	}{
		{"100", 100},
		{" 100", 100},
		{"100 ", 100},
		{" 100 ", 100},
		{"-1", -1},
	}

	for _, tt := range tests {
		data := strings.NewReader(tt.data)
		got, err := readInt(data)
		if err != nil {
			t.Errorf("readInt() should not raise error: %v", err)
		}
		if got != tt.expected {
			t.Errorf("readInt() expected %d, got %d", tt.expected, got)
		}
	}
}

func TestReadUint(t *testing.T) {
	tests := []struct {
		data     string
		expected uint64
	}{
		{"100", 100},
		{" 100", 100},
		{"100 ", 100},
		{" 100 ", 100},
	}

	for _, tt := range tests {
		data := strings.NewReader(tt.data)
		got, err := readUint(data)
		if err != nil {
			t.Errorf("readInt() should not raise error: %v", err)
		}
		if got != tt.expected {
			t.Errorf("readInt() expected %d, got %d", tt.expected, got)
		}
	}
}
