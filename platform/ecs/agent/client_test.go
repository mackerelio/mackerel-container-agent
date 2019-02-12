package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

var dockerID = "7e088b28bde202f19243853b0d20998a005984efa3d4b6c18e771fd149f86648"

func newAgentAPIServer() *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case metadataPath:
			http.ServeFile(w, r, "testdata/metadata.json")
		case taskPath:
			if r.URL.Query().Get("dockerid") == dockerID {
				http.ServeFile(w, r, "testdata/task.json")
			}
		}
	})
	return httptest.NewServer(handler)
}

func TestGetInstanceMetadata(t *testing.T) {
	ctx := context.Background()
	ts := newAgentAPIServer()
	defer ts.Close()

	c, err := NewClient(ts.URL)
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}

	_, err = c.GetInstanceMetadata(ctx)
	if err != nil {
		t.Errorf("GetInstanceMetadata() should not raise error: %v", err)
	}
}

func TestGetTaskMetadata(t *testing.T) {
	ctx := context.Background()
	ts := newAgentAPIServer()
	defer ts.Close()

	c, err := NewClient(ts.URL)
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}

	_, err = c.GetTaskMetadataWithDockerID(ctx, dockerID)
	if err != nil {
		t.Errorf("GetTaskMetadata() should not raise error: %v", err)
	}
}
