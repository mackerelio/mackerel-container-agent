package kubernetes

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/mackerelio/golib/logging"

	"github.com/mackerelio/mackerel-container-agent/metric"
	"github.com/mackerelio/mackerel-container-agent/metric/hostinfo"
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

		caCert, err = os.ReadFile(caCertificateFile)
		if err != nil {
			return nil, err
		}

		token, err = os.ReadFile(tokenFile)
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

// NewEKSOnFargatePlatform creates a new Platform
// on this platform, agent accesses Kubelet via Kubernetes API (/api/v1/nodes/{nodeName}/proxy)
func NewEKSOnFargatePlatform(kubernetesHost, kubernetesPort string, namespace, podName string, nodeName string, ignoreContainer *regexp.Regexp) (platform.Platform, error) {
	var caCert, token []byte
	var err error

	baseURL := &url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort(kubernetesHost, kubernetesPort),
		Path:   path.Join("api", "v1", "nodes", nodeName, "proxy"),
	}

	caCert, err = os.ReadFile(caCertificateFile)
	if err != nil {
		return nil, err
	}

	token, err = os.ReadFile(tokenFile)
	if err != nil {
		return nil, err
	}

	httpClient := createHTTPClient(caCert, false)

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
		newMetricGenerator(p.client, hostinfo.NewGenerator()),
		metric.NewInterfaceGenerator(),
	}
}

// GetSpecGenerators gets spec generator
func (p *kubernetesPlatform) GetSpecGenerators() []spec.Generator {
	return []spec.Generator{
		newSpecGenerator(p.client),
		&spec.CPUGenerator{},
	}
}

// GetCustomIdentifier gets custom identifier
func (p *kubernetesPlatform) GetCustomIdentifier(ctx context.Context) (string, error) {
	pod, err := p.client.GetPod(ctx)
	if err != nil {
		return "", err
	}
	return string(pod.UID) + ".kubernetes", nil
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
	dt := http.DefaultTransport.(*http.Transport)
	tp := &http.Transport{
		Proxy:                 nil,
		DialContext:           dt.DialContext,
		MaxIdleConns:          dt.MaxIdleConns,
		IdleConnTimeout:       dt.IdleConnTimeout,
		TLSHandshakeTimeout:   dt.TLSHandshakeTimeout,
		ExpectContinueTimeout: dt.ExpectContinueTimeout,
	}

	tp.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: insecureTLS,
	}

	if len(caCert) > 0 {
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(caCert)
		tp.TLSClientConfig.RootCAs = certPool
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: tp,
	}
}
