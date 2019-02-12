package instance

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"
)

func newInstanceMetadataAPIServer() *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == path.Join(basePath, localIPv4Path) {
			w.Write([]byte("10.0.1.100"))
		}
	})
	return httptest.NewServer(handler)
}

func TestGetLocalIPv4(t *testing.T) {
	ctx := context.Background()
	ts := newInstanceMetadataAPIServer()
	defer ts.Close()

	c, err := NewClient(ts.URL)
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}

	addr, err := c.GetLocalIPv4(ctx)
	if err != nil {
		t.Errorf("GetLocalIPv4() should not raise error: %v", err)
	}
	if addr != "10.0.1.100" {
		t.Errorf("GetLocalIPv4() expected 10.0.1.100, got %s", addr)
	}
}
