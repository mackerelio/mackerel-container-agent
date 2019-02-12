package probe

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/mackerelio/mackerel-container-agent/config"
)

func init() {
	defaultTimeoutHTTP = 100 * time.Millisecond
}

func TestProbeHTTP_Check(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		path           string
		headers        []config.Header
		timeoutSeconds int
		status         int
		sleep          time.Duration
		shouldErr      bool
	}{
		{
			name:   "ok",
			path:   "/healthy",
			status: http.StatusOK,
		},
		{
			name:   "3xx",
			path:   "/healthy",
			status: http.StatusNotModified,
		},
		{
			name:      "4xx",
			path:      "/healthy",
			status:    http.StatusBadRequest,
			shouldErr: true,
		},
		{
			name:      "5xx",
			path:      "/healthy",
			status:    http.StatusServiceUnavailable,
			shouldErr: true,
		},
		{
			name:   "5xx",
			method: "PUT",
			path:   "/healthy",
			status: http.StatusOK,
		},
		{
			name:      "timeout",
			path:      "/healthy",
			status:    http.StatusOK,
			sleep:     300 * time.Millisecond,
			shouldErr: true,
		},
		{
			name:           "timeout seconds",
			path:           "/healthy",
			timeoutSeconds: 1,
			status:         http.StatusOK,
			sleep:          300 * time.Millisecond,
		},
		{
			name:    "headers",
			path:    "/healthy",
			headers: []config.Header{{Name: "X-Custom-Header", Value: "test"}},
			status:  http.StatusOK,
		},
		{
			name:    "host",
			path:    "/healthy",
			headers: []config.Header{{Name: "Host", Value: "example.com"}},
			status:  http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := newHTTPServer(t, "ok", tc.headers, tc.method, tc.path, tc.sleep, tc.status)
			defer ts.Close()
			u, _ := url.Parse(ts.URL)

			p := NewProbe(&config.Probe{
				HTTP: &config.ProbeHTTP{
					Scheme:  u.Scheme,
					Host:    u.Hostname(),
					Port:    u.Port(),
					Method:  tc.method,
					Path:    tc.path,
					Headers: tc.headers,
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

func newHTTPServer(t *testing.T, content string, headers []config.Header, method, path string, sleep time.Duration, status int) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if method != "" && r.Method != method {
			t.Errorf("method should be %s but got %s", method, r.Method)
		}
		if r.URL.Path != path {
			t.Errorf("path should be %s but got %s", path, r.URL.Path)
		}
		for _, h := range headers {
			if strings.ToLower(h.Name) == "host" {
				if expected := h.Value; r.Host != expected {
					t.Errorf("host should be %s but got %s", expected, r.Host)
				}
				continue
			}
			if expected := h.Value; r.Header.Get(h.Name) != expected {
				t.Errorf("header %s should set %s but got %s", h.Name, expected, r.Header.Get(h.Name))
			}
		}
		time.Sleep(sleep)
		w.WriteHeader(status)
		w.Write([]byte(content))
	})
	return httptest.NewServer(handler)
}
