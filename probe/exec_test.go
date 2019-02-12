package probe

import (
	"context"
	"testing"
	"time"

	"github.com/mackerelio/mackerel-container-agent/cmdutil"
	"github.com/mackerelio/mackerel-container-agent/config"
)

func init() {
	defaultTimeoutExec = 100 * time.Millisecond
}

func TestProbeExec_Check(t *testing.T) {
	testCases := []struct {
		name           string
		command        string
		timeoutSeconds int
		env            []string
		shouldErr      bool
	}{
		{
			name:    "ok",
			command: "echo ok",
		},
		{
			name:    "exit 0",
			command: "exit 0",
		},
		{
			name:      "exit 1",
			command:   "exit 1",
			shouldErr: true,
		},
		{
			name:      "timeout",
			command:   "sleep 5",
			shouldErr: true,
		},
		{
			name:           "timeout seconds",
			command:        "sleep 0.5",
			timeoutSeconds: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewProbe(&config.Probe{
				Exec: &config.ProbeExec{
					Command: cmdutil.CommandString(tc.command),
					Env:     tc.env,
				},
				TimeoutSeconds: tc.timeoutSeconds,
			})
			err := p.Check(context.Background())

			if err != nil && !tc.shouldErr {
				t.Errorf("should not raise error: %v", err)
			}
			if err == nil && tc.shouldErr {
				t.Errorf("should raise error: %v", err)
			}
		})
	}
}
