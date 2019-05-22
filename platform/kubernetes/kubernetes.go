package kubernetes

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/mackerelio/golib/logging"

	"github.com/mackerelio/mackerel-container-agent/metric"
	"github.com/mackerelio/mackerel-container-agent/platform"
	"github.com/mackerelio/mackerel-container-agent/platform/kubernetes/kubelet"
	"github.com/mackerelio/mackerel-container-agent/spec"
)

var (
	logger  = logging.GetLogger("kubernetes")
	timeout = 3 * time.Second

	caCertificateFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	tokenFile         = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

type kubernetesPlatform struct {
	client kubelet.Client
}

// NewKubernetesPlatform creates a new Platform
func NewKubernetesPlatform(kubeletHost, kubeletPort string, useReadOnlyPort, insecureTLS bool, namespace, podName string, ignoreContainer *regexp.Regexp) (platform.Platform, error) {
	var caCert, token []byte

	baseURL := &url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(kubeletHost, kubeletPort),
	}

	if !useReadOnlyPort {
		baseURL.Scheme = "https"

		var err error

		caCert, err = ioutil.ReadFile(caCertificateFile)
		if err != nil {
			return nil, err
		}

		token, err = ioutil.ReadFile(tokenFile)
		if err != nil {
			return nil, err
		}
	}

	httpClient := createHTTPClient(caCert, insecureTLS)

	c, err := kubelet.NewClient(
		httpClient,
		string(token),
		baseURL.String(),
		namespace,
		podName,
		ignoreContainer,
	)
	if err != nil {
		return nil, err
	}
	return &kubernetesPlatform{client: c}, nil
}

// GetMetricGenerators gets metric generators
func (p *kubernetesPlatform) GetMetricGenerators() []metric.Generator {
	return []metric.Generator{
		newMetricGenerator(p.client),
		metric.NewInterfaceGenerator(),
	}
}

// GetSpecGenerators gets spec generator
func (p *kubernetesPlatform) GetSpecGenerators() []spec.Generator {
	return []spec.Generator{
		newSpecGenerator(p.client),
	}
}

// GetCustomIdentifier gets custom identifier
func (p *kubernetesPlatform) GetCustomIdentifier(ctx context.Context) (string, error) {
	pod, err := p.client.GetPod(ctx)
	if err != nil {
		return "", err
	}
	return string(pod.ObjectMeta.UID) + ".kubernetes", nil
}

// StatusRunning reports p status is running
func (p *kubernetesPlatform) StatusRunning(ctx context.Context) bool {
	meta, err := p.client.GetPod(ctx)
	if err != nil {
		logger.Warningf("failed to get metadata: %s", err)
		return false
	}
	return strings.EqualFold("running", string(meta.Status.Phase))
}

func createHTTPClient(caCert []byte, insecureTLS bool) *http.Client {
	// Copy from the definition of http.DefaultTransport.DialContext.
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}
	// Copy from the definition of http.DefaultTransport.
	// Don't use Proxy.
	transport := &http.Transport{
		Dial:                  dialer.Dial,
		DialContext:           dialer.DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: insecureTLS,
	}

	if len(caCert) > 0 {
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(caCert)
		transport.TLSClientConfig.RootCAs = certPool
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}
