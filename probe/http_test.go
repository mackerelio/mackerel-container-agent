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
		useProxy       bool
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
		{
			name:     "proxy",
			path:     "/",
			status:   http.StatusOK,
			useProxy: true,
		},
	}

	var proxyHandler func(*http.Request) (*url.URL, error)

	dt := http.DefaultTransport.(*http.Transport)
	origProxy := dt.Proxy
	defer func() {
		dt.Proxy = origProxy
	}()
	dt.Proxy = func(req *http.Request) (*url.URL, error) {
		return proxyHandler(req)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := newHTTPServer(t, "ok", tc.headers, tc.method, tc.path, tc.sleep, tc.status)
			u, _ := url.Parse(ts.URL)

			var passedProxy bool
			proxyHandler = func(req *http.Request) (*url.URL, error) {
				passedProxy = true
				return nil, nil
			}

			var proxyURL *url.URL
			if tc.useProxy {
				ps := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					passedProxy = true
					ts.Config.Handler.ServeHTTP(w, r)
				}))
				proxyURL, _ = url.Parse(ps.URL)
				defer ps.Close()
			}

			p := NewProbe(&config.Probe{
				HTTP: &config.ProbeHTTP{
					Scheme:  u.Scheme,
					Host:    u.Hostname(),
					Port:    u.Port(),
					Method:  tc.method,
					Path:    tc.path,
					Headers: tc.headers,
					Proxy:   config.URLWrapper{URL: proxyURL},
				},
				TimeoutSeconds: tc.timeoutSeconds,
			})

			err := p.Check(context.Background())

			if tc.useProxy && !passedProxy {
				t.Errorf("request should through the proxy")
			}
			if !tc.useProxy && passedProxy {
				t.Errorf("request should not through the proxy")
			}

			if err != nil && !tc.shouldErr {
				t.Errorf("should not raise error: %v", err)
			}
			if err == nil && tc.shouldErr {
				t.Errorf("should raise error: %v", err)
			}
		})
	}
}

func newHTTPServer(t testing.TB, content string, headers []config.Header, method, path string, sleep time.Duration, status int) *httptest.Server {
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
		// Below writing a content sometimes expects returning an error.
		// A HTTP client may disconnect from the server during above time.Sleep.
		w.Write([]byte(content)) // nolint
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return ts
}
