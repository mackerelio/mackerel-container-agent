package kubernetesapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"testing"
)

func newServer(token string) *httptest.Server {
	validNodePath := path.Join(basePath, nodePath, "VALID-NODE")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token != "" && r.Header.Get("Authorization") != "Bearer "+token {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		fmt.Println(r.URL.RequestURI())
		// only /api/v1/nodes/VALID-NODE is valid
		switch r.URL.RequestURI() {
		case validNodePath:
			http.ServeFile(w, r, "testdata/node.json")
		}
	})
	return httptest.NewServer(handler)
}

func TestGetNode(t *testing.T) {
	ctx := context.Background()
	ts := newServer("")
	defer ts.Close()
	c, err := NewClient(ts.Client(), "", ts.URL, "VALID-NODE")
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	_, err = c.GetNode(ctx)
	if err != nil {
		t.Errorf("GetNode() should not raise error: %v", err)
	}
}

func TestRequestToken(t *testing.T) {
	testToken := "testToken"

	ts := newServer(testToken)
	defer ts.Close()

	ctx := context.Background()
	c, err := NewClient(ts.Client(), testToken, ts.URL, "VALID-NODE")
	if err != nil {
		t.Errorf("NewClient() should not raise error: %v", err)
	}

	_, err = c.GetNode(ctx)
	if err != nil {
		t.Errorf("newRequest() should not raise error: %v", err)
	}
}

func TestErrorMessage(t *testing.T) {
	var body string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(body))
	})
	ts := httptest.NewServer(handler)

	c, err := NewClient(ts.Client(), "", ts.URL, "some-node")
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

		_, err = c.GetNode(ctx)
		if err == nil {
			t.Errorf("should raise error")
		}

		u, _ := url.Parse(ts.URL)
		u.Path = path.Join(u.Path, basePath, nodePath, "some-node")
		expected := fmt.Sprintf("got status code %d (url: %s, body: %q)", http.StatusBadRequest, u, tc.body)

		got := err.Error()
		if got != expected {
			t.Errorf("error message expected %q, got %q", expected, got)
		}
	}
}
