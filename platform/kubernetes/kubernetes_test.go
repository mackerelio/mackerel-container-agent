package kubernetes

import (
	"bytes"
	"context"
	"encoding/pem"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mackerelio/mackerel-container-agent/platform/kubernetes/kubelet"
)

func TestStatusRunning(t *testing.T) {
	mockClient := kubelet.NewMockClient()
	pform := kubernetesPlatform{mockClient}

	tests := []struct {
		status string
		expect bool
	}{
		{"running", true},
		{"Running", true},
		{"RUNNING", true},
		{"PENDING", false},
		{"", false},
	}

	for _, tc := range tests {
		ctx := context.Background()
		mockClient.ApplyOption(
			kubelet.MockGetPod(
				func(context.Context) (*kubelet.Pod, error) {
					return &kubelet.Pod{
						Status: kubelet.PodStatus{Phase: tc.status},
					}, nil
				},
			),
		)

		got := pform.StatusRunning(ctx)
		if got != tc.expect {
			t.Errorf("StatusRunning() expected %t, got %t", tc.expect, got)
		}
	}
}

func TestCreateHTTPClient(t *testing.T) {
	host := "example.com"
	body := "hello world"
	var addr, port string

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := net.JoinHostPort(host, port)
		if r.Host != expected {
			t.Errorf("requested host expected %q, but %q", expected, r.Host)
		}
		w.Write([]byte(body))
	}))

	addr, port, _ = net.SplitHostPort(ts.Listener.Addr().String())
	rslv := &resolver{
		host:    host,
		address: addr,
		port:    port,
	}

	caCert := &bytes.Buffer{}
	pem.Encode(caCert, &pem.Block{Type: "CERTIFICATE", Bytes: ts.Certificate().Raw})

	client := createHTTPClient(caCert.Bytes(), rslv)
	resp, err := client.Get("https://" + net.JoinHostPort(host, port))
	if err != nil {
		t.Errorf("Get() should not raise error: %v", err)
	}

	got, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if string(got) != body {
		t.Errorf("response body expected %q, got %q", body, got)
	}
}
