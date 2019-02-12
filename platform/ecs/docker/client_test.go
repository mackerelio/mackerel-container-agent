package docker

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"
)

var containerID = "4a81836ecac41f16c09e1f2fd39b35162718c3726f3651af9ea4744a1d0d79a3"

func newDockerAPIServerOnUnixDomainSocket() (*httptest.Server, error) {
	tempDir, err := ioutil.TempDir("", "dockertapi-server")
	if err != nil {
		return nil, err
	}
	listener, err := net.Listen("unix", path.Join(tempDir, "server.sock"))
	if err != nil {
		return nil, err
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.String() {
		case fmt.Sprintf(statsPath+"?stream=false", containerID):
			http.ServeFile(w, r, fmt.Sprintf("testdata/%s/stats.json", containerID))
		case fmt.Sprintf(inspectPath, containerID):
			http.ServeFile(w, r, fmt.Sprintf("testdata/%s/inspect.json", containerID))
		}
	})
	server := &httptest.Server{
		Listener: listener,
		Config:   &http.Server{Handler: handler},
	}
	server.Start()
	return server, nil
}

func TestGetContainerStats(t *testing.T) {
	ctx := context.Background()
	ts, err := newDockerAPIServerOnUnixDomainSocket()
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	defer ts.Close()
	sock := ts.Listener.Addr()
	c, err := NewClient(sock.String())
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	stats, err := c.GetContainerStats(ctx, containerID)
	if err != nil {
		t.Errorf("GetContainerStats() should not raise error: %v", err)
	}
	if stats.ID != containerID {
		t.Errorf("stats.ID expected %s, got %s", containerID, stats.ID)
	}
}

func TestInspectContainer(t *testing.T) {
	ctx := context.Background()
	ts, err := newDockerAPIServerOnUnixDomainSocket()
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	defer ts.Close()
	sock := ts.Listener.Addr()
	c, err := NewClient(sock.String())
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	container, err := c.InspectContainer(ctx, containerID)
	if err != nil {
		t.Errorf("InspectContainer() should not raise error: %v", err)
	}
	if container.ID != containerID {
		t.Errorf("container.ID expected %s, got %s", containerID, container.ID)
	}
}
