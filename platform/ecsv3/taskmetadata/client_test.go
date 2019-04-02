package taskmetadata

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func newServer() *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.RequestURI() {
		case metadataPath:
			http.ServeFile(w, r, "testdata/metadata_ec2_awsvpc.json")
			// case statsPath:
			//   http.ServeFile(w, r, "testdata/stats.json")
		}
	})
	return httptest.NewServer(handler)
}

func TestGetTaskMetadata(t *testing.T) {
	ctx := context.Background()
	ts := newServer()
	defer ts.Close()

	c, err := NewClient(ts.URL, nil)
	if err != nil {
		t.Errorf("NewClient() should not raise error: %v", err)
	}

	_, err = c.GetTaskMetadata(ctx)
	if err != nil {
		t.Errorf("GetTaskMetadata() should not raise error: %v", err)
	}
}

// func TestGetTaskStats(t *testing.T) {
//   ctx := context.Background()
//   ts := newServer()
//   defer ts.Close()

//   c, err := NewClient(ts.URL, nil)
//   if err != nil {
//     t.Errorf("NewClient() should not raise error: %v", err)
//   }

//   _, err = c.GetTaskStats(ctx)
//   if err != nil {
//     t.Errorf("GetTaskStats() should not raise error: %v", err)
//   }
// }

func TestIgnoreContainer(t *testing.T) {
	ts := newServer()
	defer ts.Close()

	tests := []struct {
		ignoreContainer *regexp.Regexp
		expected        int
	}{
		{nil, 2},
		{regexp.MustCompile(`\A~internal~ecs~pause\z`), 1},
		{regexp.MustCompile(``), 0},
	}

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

		// stats, err := c.GetStats(ctx)
		// if err != nil {
		//   t.Errorf("GetStats() should not raise error: %v", err)
		// }
		// got = len(stats)
		// if got != tc.expected {
		//   t.Errorf("GetStats() expected %d containers, got %v containers", tc.expected, got)
		// }
	}

}
