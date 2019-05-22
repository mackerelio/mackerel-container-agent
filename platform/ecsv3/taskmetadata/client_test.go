package taskmetadata

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"regexp"
	"testing"
)

var metadata, stats string

func newServer() *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.RequestURI() {
		case metadataPath:
			http.ServeFile(w, r, metadata)
		case statsPath:
			http.ServeFile(w, r, stats)
		}
	})
	return httptest.NewServer(handler)
}

func TestGetTaskMetadata(t *testing.T) {
	tests := []string{
		"testdata/metadata_ec2_bridge.json",
		"testdata/metadata_ec2_host.json",
		"testdata/metadata_ec2_awsvpc.json",
		"testdata/metadata_fargate.json",
	}

	ts := newServer()
	defer ts.Close()

	for _, path := range tests {
		metadata = path
		ctx := context.Background()

		c, err := NewClient(ts.URL, nil)
		if err != nil {
			t.Errorf("NewClient() should not raise error: %v", err)
		}

		_, err = c.GetTaskMetadata(ctx)
		if err != nil {
			t.Errorf("GetTaskMetadata() should not raise error: %v", err)
		}
	}

}

func TestGetTaskStats(t *testing.T) {
	tests := []string{
		"testdata/stats_ec2_bridge.json",
		"testdata/stats_ec2_host.json",
		"testdata/stats_ec2_awsvpc.json",
		"testdata/stats_fargate.json",
	}

	ts := newServer()
	defer ts.Close()

	for _, path := range tests {
		stats = path
		ctx := context.Background()

		c, err := NewClient(ts.URL, nil)
		if err != nil {
			t.Errorf("NewClient() should not raise error: %v", err)
		}

		_, err = c.GetTaskStats(ctx)
		if err != nil {
			t.Errorf("GetTaskStats() should not raise error: %v", err)
		}
	}
}

func TestIgnoreContainer(t *testing.T) {
	tests := []struct {
		ignoreContainer *regexp.Regexp
		expected        int
	}{
		{nil, 2},
		{regexp.MustCompile(`\A~internal~ecs~pause\z`), 1},
		{regexp.MustCompile(``), 0},
	}

	ts := newServer()
	defer ts.Close()

	metadata = "testdata/metadata_ec2_awsvpc.json"
	stats = "testdata/stats_ec2_awsvpc.json"

	for _, tc := range tests {
		ctx := context.Background()

		c, err := NewClient(ts.URL, tc.ignoreContainer)
		if err != nil {
			t.Errorf("should not raise error: %v", err)
		}

		meta, err := c.GetTaskMetadata(ctx)
		if err != nil {
			t.Errorf("GetTaskMetadata() should not raise error: %v", err)
		}
		got := len(meta.Containers)
		if got != tc.expected {
			t.Errorf("meta.Containers expected %d containers, got %v containers", tc.expected, got)
		}

		stats, err := c.GetTaskStats(ctx)
		if err != nil {
			t.Errorf("GetStats() should not raise error: %v", err)
		}
		got = len(stats)
		if got != tc.expected {
			t.Errorf("GetStats() expected %d containers, got %v containers", tc.expected, got)
		}
	}

}

func TestErrorMessage(t *testing.T) {
	var body string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(body))
	})
	ts := httptest.NewServer(handler)

	c, err := NewClient(ts.URL, nil)
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}

	tests := []struct {
		body string
	}{
		{"Bad Request"},
		{"Bad\nRequest"},
	}

	for _, tc := range tests {
		body = tc.body

		ctx := context.Background()

		_, err = c.GetTaskMetadata(ctx)
		if err == nil {
			t.Errorf("should raise error")
		}

		u, _ := url.Parse(ts.URL)
		u.Path = path.Join(u.Path, metadataPath)
		expected := fmt.Sprintf("got status code %d (url: %s, body: %q)", http.StatusBadRequest, u, tc.body)

		got := err.Error()
		if got != expected {
			t.Errorf("error message expected %q, got %q", expected, got)
		}
	}
}

func TestNoProxy(t *testing.T) {
	var useProxy bool

	dt := http.DefaultTransport.(*http.Transport)
	origProxy := dt.Proxy
	defer func() {
		dt.Proxy = origProxy
	}()
	dt.Proxy = func(req *http.Request) (*url.URL, error) {
		useProxy = true
		return nil, nil
	}

	th := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	ts := httptest.NewServer(th)

	c, _ := NewClient(ts.URL, nil)
	c.GetTaskMetadata(context.Background())

	if useProxy == true {
		t.Error("proxy should not be used")
	}
}
