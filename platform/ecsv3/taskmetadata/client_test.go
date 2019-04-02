package taskmetadata

import (
	"context"
	"net/http"
	"net/http/httptest"
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
