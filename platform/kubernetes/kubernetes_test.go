package kubernetes

import (
	"bytes"
	"context"
	"encoding/pem"
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
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	host, port, _ := net.SplitHostPort(ts.Listener.Addr().String())

	caCert := &bytes.Buffer{}
	pem.Encode(caCert, &pem.Block{Type: "CERTIFICATE", Bytes: ts.Certificate().Raw})

	tests := []struct {
		caCert      []byte
		insecureTLS bool
		expect      bool
	}{
		{caCert.Bytes(), false, true},
		{caCert.Bytes(), true, true},
		{[]byte{}, false, false},
		{[]byte{}, true, true},
	}

	url := "https://" + net.JoinHostPort(host, port)

	for _, tc := range tests {
		client := createHTTPClient(tc.caCert, tc.insecureTLS)
		resp, err := client.Get(url)
		if (err == nil) != tc.expect {
			t.Errorf("Get() does not expected benavior: %v", err)
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
}
