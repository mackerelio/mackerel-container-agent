package ecsv3

import (
	"testing"
)

func TestIsRunning(t *testing.T) {
	tests := []struct {
		status string
		expect bool
	}{
		{"running", true},
		{"Running", true},
		{"RUNNING", true},
		{"PENDING", false},
		{"", false},
	}

	for _, tc := range tests {
		got := isRunning(tc.status)
		if got != tc.expect {
			t.Errorf("isRunning() expected %t, got %t", tc.expect, got)
		}
	}
}
