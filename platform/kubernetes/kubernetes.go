package kubernetes

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/mackerelio/golib/logging"

	"github.com/mackerelio/mackerel-container-agent/metric"
	"github.com/mackerelio/mackerel-container-agent/platform"
	"github.com/mackerelio/mackerel-container-agent/platform/kubernetes/kubelet"
	"github.com/mackerelio/mackerel-container-agent/platform/kubernetes/kubernetesapi"
	"github.com/mackerelio/mackerel-container-agent/spec"
)

var (
	logger  = logging.GetLogger("kubernetes")
	timeout = 3 * time.Second

	caCertificateFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	tokenFile         = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

type kubernetesPlatform struct {
	client    kubelet.Client // TODO rename to something like `kubeletClient` ?
	apiClient kubernetesapi.Client
}

// NewKubernetesPlatform creates a new Platform
func NewKubernetesPlatform(kubernetesServiceHost, kubernetesServicePort, kubeletHost, kubeletPort string, useReadOnlyPort, insecureTLS bool, namespace, podName string, nodeName string, ignoreContainer *regexp.Regexp) (platform.Platform, error) {
	var caCert, token []byte
	var err error

	baseURL := &url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort(kubernetesServiceHost, kubernetesServicePort),
	}

	kubeletBaseURL := &url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(kubeletHost, kubeletPort),
	}

	if !useReadOnlyPort {
		kubeletBaseURL.Scheme = "https"
	}

	caCert, err = ioutil.ReadFile(caCertificateFile)
	if err != nil {
		return nil, err
	}

	token, err = ioutil.ReadFile(tokenFile)
	if err != nil {
		return nil, err
	}

	httpClient := createHTTPClient(caCert, insecureTLS)

	sc, err := kubernetesapi.NewClient(
		httpClient,
		string(token),
		baseURL.String(),
		nodeName,
	)
	if err != nil {
		return nil, err
	}
	c, err := kubelet.NewClient(
		httpClient,
		string(token),
		kubeletBaseURL.String(),
		namespace,
		podName,
		ignoreContainer,
	)
	if err != nil {
		return nil, err
	}
	return &kubernetesPlatform{client: c, apiClient: sc}, nil
}

// NewEKSOnFargatePlatform creates a new Platform
func NewEKSOnFargatePlatform(kubernetesServiceHost, kubernetesServicePort string, namespace, podName string, nodeName string, ignoreContainer *regexp.Regexp) (platform.Platform, error) {
	var caCert, token []byte
	var err error

	baseURL := &url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort(kubernetesServiceHost, kubernetesServicePort),
	}

	// Access to kubelet via /api/v1/nodes/{nodeName}/proxy
	kubeletBaseURL := &url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort(kubernetesServiceHost, kubernetesServicePort),
		Path:   path.Join("api", "v1", "nodes", nodeName, "proxy"),
	}

	caCert, err = ioutil.ReadFile(caCertificateFile)
	if err != nil {
		return nil, err
	}

	token, err = ioutil.ReadFile(tokenFile)
	if err != nil {
		return nil, err
	}

	httpClient := createHTTPClient(caCert, false)

	sc, err := kubernetesapi.NewClient(
		httpClient,
		string(token),
		baseURL.String(),
		nodeName,
	)
	if err != nil {
		return nil, err
	}

	c, err := kubelet.NewClient(
		httpClient,
		string(token),
		kubeletBaseURL.String(),
		namespace,
		podName,
		ignoreContainer,
	)
	if err != nil {
		return nil, err
	}
	return &kubernetesPlatform{client: c, apiClient: sc}, nil
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
		&spec.CPUGenerator{},
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
