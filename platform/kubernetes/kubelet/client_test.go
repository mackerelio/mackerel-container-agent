package kubelet

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func newServer(token string) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token != "" && r.Header.Get("Authorization") != "Bearer "+token {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch r.URL.RequestURI() {
		case podsPath:
			http.ServeFile(w, r, "testdata/pods.json")
		case statsPath:
			http.ServeFile(w, r, "testdata/summary.json")
		case specPath:
			http.ServeFile(w, r, "testdata/spec.json")
		case "/spec":
			w.Header().Set("Content-Type", "text/html")
			w.Header().Set("location", specPath)
			w.WriteHeader(http.StatusMovedPermanently)
		}
	})
	return httptest.NewServer(handler)
}

func TestGetPod(t *testing.T) {
	ts := newServer("")
	defer ts.Close()

	tests := []struct {
		namespace, name string
		raiseError      bool
	}{
		{
			namespace:  "default",
			name:       "myapp",
			raiseError: false,
		},
		{
			namespace:  "default",
			name:       "dummy",
			raiseError: true,
		},
		{
			namespace:  "dummy",
			name:       "myapp",
			raiseError: true,
		},
		{
			namespace:  "dummy",
			name:       "dummy",
			raiseError: true,
		},
		{
			namespace:  "default",
			name:       "",
			raiseError: true,
		},
		{
			namespace:  "",
			name:       "myapp",
			raiseError: true,
		},
		{
			namespace:  "",
			name:       "",
			raiseError: true,
		},
	}

	for _, tc := range tests {
		ctx := context.Background()
		c, err := NewClient(ts.Client(), "", ts.URL, tc.namespace, tc.name, nil)
		if err != nil {
			t.Errorf("should not raise error: %v", err)
		}

		_, err = c.GetPod(ctx)
		if (err != nil) != tc.raiseError {
			var cond string
			if !tc.raiseError {
				cond = "not "
			}
			t.Errorf("GetPod() should %sraise error, but got %q", cond, err)
		}
	}
}

func TestGetPodStats(t *testing.T) {
	ts := newServer("")
	defer ts.Close()

	tests := []struct {
		namespace, name string
		raiseError      bool
	}{
		{
			namespace:  "default",
			name:       "myapp",
			raiseError: false,
		},
		{
			namespace:  "default",
			name:       "dummy",
			raiseError: true,
		},
		{
			namespace:  "dummy",
			name:       "myapp",
			raiseError: true,
		},
		{
			namespace:  "dummy",
			name:       "dummy",
			raiseError: true,
		},
		{
			namespace:  "default",
			name:       "",
			raiseError: true,
		},
		{
			namespace:  "",
			name:       "myapp",
			raiseError: true,
		},
		{
			namespace:  "",
			name:       "",
			raiseError: true,
		},
	}

	for _, tc := range tests {
		ctx := context.Background()
		c, err := NewClient(ts.Client(), "", ts.URL, tc.namespace, tc.name, nil)
		if err != nil {
			t.Errorf("should not raise error: %v", err)
		}

		_, err = c.GetPodStats(ctx)
		if (err != nil) != tc.raiseError {
			var cond string
			if !tc.raiseError {
				cond = "not "
			}
			t.Errorf("GetPodStats() should %sraise error, but got %q", cond, err)
		}
	}
}

func TestGetSpec(t *testing.T) {
	ctx := context.Background()
	ts := newServer("")
	defer ts.Close()
	c, err := NewClient(ts.Client(), "", ts.URL, "default", "myapp", nil)
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	_, err = c.GetSpec(ctx)
	if err != nil {
		t.Errorf("GetSpec() should not raise error: %v", err)
	}
}

func TestIgnoreContainer(t *testing.T) {
	ts := newServer("")
	defer ts.Close()

	tests := []struct {
		ignoreContainer *regexp.Regexp
		expected        int
	}{
		{nil, 2},
		{regexp.MustCompile(`\Amackerel-container-agent\z`), 1},
		{regexp.MustCompile(``), 0},
	}

	for _, tc := range tests {
		ctx := context.Background()
		c, err := NewClient(ts.Client(), "", ts.URL, "default", "myapp", tc.ignoreContainer)
		if err != nil {
			t.Errorf("should not raise error: %v", err)
		}

		pod, err := c.GetPod(ctx)
		if err != nil {
			t.Errorf("GetPod() should not raise error: %v", err)
		}
		got := len(pod.Spec.Containers)
		if got != tc.expected {
			t.Errorf("meta.Containers expected %d containers, got %v containers", tc.expected, got)
		}

		stats, err := c.GetPodStats(ctx)
		if err != nil {
			t.Errorf("GetPodStats() should not raise error: %v", err)
		}
		got = len(stats.Containers)
		if got != tc.expected {
			t.Errorf("GetPodStats() expected %d containers, got %v containers", tc.expected, got)
		}
	}

}

func TestRequestToken(t *testing.T) {
	testToken := "testToken"

	ts := newServer(testToken)
	defer ts.Close()

	ctx := context.Background()
	c, err := NewClient(ts.Client(), testToken, ts.URL, "default", "myapp", nil)
	// c, err := NewClient(ts.Client(), "hoge", ts.URL, "default", "myapp", nil)
	if err != nil {
		t.Errorf("NewClient() should not raise error: %v", err)
	}

	_, err = c.GetSpec(ctx)
	if err != nil {
		t.Errorf("newRequest() should not raise error: %v", err)
	}
}
