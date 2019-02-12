package probe

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/mackerelio/mackerel-container-agent/config"
)

func init() {
	defaultTimeoutTCP = 100 * time.Millisecond
}

func TestProbeTCP_Check(t *testing.T) {
	testCases := []struct {
		name      string
		port      string
		shouldErr bool
	}{
		{
			name: "ok",
		},
		{
			name:      "invalid port",
			port:      "1",
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := newHTTPServer(t, "ok", nil, "GET", "/", 0, http.StatusOK)
			defer ts.Close()
			u, _ := url.Parse(ts.URL)

			port := u.Port()
			if tc.port != "" {
				port = tc.port
			}
			p := NewProbe(&config.Probe{
				TCP: &config.ProbeTCP{
					Host: u.Hostname(),
					Port: port,
				},
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
