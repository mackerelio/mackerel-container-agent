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

	// caCert and token are:
	// - necessary when useReadOnlyPort == true with Kubernetes 1.18+
	//   - in this case we use them to access Kubernetes API
	// - NOT necessary when useReadOnlyPort == true with Kubernets < 1.18
	//   - in this case we don't access Kubernetes API
	// - necessary when useReadOnlyPort == false, regardless Kubernetes version
	//   - in this case we use them to access Kubelet
	//   - additionaly, on Kubernetes 1.18+ they are also used to access Kubernetes API
	// therefore:
	// - will return err when we cannot load them and useReadOnlyPort == true
	// - will NOT return err when we cannot load them and useReadOnlyPort == false
	//   - instead we just log
	// TODO: return err even useReadOnlyPort after we drop Kubernetes < 1.18
	caCert, err = ioutil.ReadFile(caCertificateFile)
	if err != nil {
		if !useReadOnlyPort {
			return nil, err
		}
		logger.Warningf("failed to read service account certification file, %w", err)
	}

	token, err = ioutil.ReadFile(tokenFile)
	if err != nil {
		if !useReadOnlyPort {
			return nil, err
		}
		logger.Warningf("failed to read service account token file, %w", err)
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

	// access kubelet via proxy
	kubeletBaseURL := sc.NodeProxyURL()
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
		newMetricGenerator(p.client, p.apiClient),
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
