package kubelet

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"

	cadvisorTypes "github.com/google/cadvisor/info/v1"
)

// Client interface gets metadata and stats
type Client interface {
	GetPod(context.Context) (*Pod, error)
	GetPodStats(context.Context) (*PodStats, error)
	GetSpec(context.Context) (*cadvisorTypes.MachineInfo, error)
}

const (
	// DefaultPort represents Kubelet port
	DefaultPort = "10250"
	// DefaultReadOnlyPort represents Kubelet read-only port
	DefaultReadOnlyPort = "10255"

	podsPath  = "/pods"
	statsPath = "/stats/summary"
	specPath  = "/spec/"
)

type client struct {
	url             *url.URL
	httpClient      *http.Client
	namespace       string
	name            string
	token           string
	ignoreContainer *regexp.Regexp
}

// NewClient creates a new Client
func NewClient(httpClient *http.Client, token, baseURL, namespace, name string, ignoreContainer *regexp.Regexp) (Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	return &client{
		url:             u,
		namespace:       namespace,
		name:            name,
		httpClient:      httpClient,
		token:           token,
		ignoreContainer: ignoreContainer,
	}, nil
}

// GetPod gets pod
func (c *client) GetPod(ctx context.Context) (*Pod, error) {
	req, err := c.newRequest(podsPath)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	var podList PodList
	if err = decodeBody(resp, &podList); err != nil {
		return nil, err
	}

	var pod *Pod
	for _, p := range podList.Items {
		if p.Metadata.Namespace == c.namespace && p.Metadata.Name == c.name {
			pod = p
			break
		}
	}
	if pod == nil {
		return nil, fmt.Errorf("pod %s.%s not found", c.namespace, c.name)
	}

	if c.ignoreContainer != nil {
		containers := make([]Container, 0, len(pod.Spec.Containers))
		for _, container := range pod.Spec.Containers {
			if c.ignoreContainer.MatchString(container.Name) {
				continue
			}
			containers = append(containers, container)
		}
		pod.Spec.Containers = containers
	}

	return pod, nil
}

// GetPodStats gets pod stats
func (c *client) GetPodStats(ctx context.Context) (*PodStats, error) {
	req, err := c.newRequest(statsPath)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	var summary Summary
	if err = decodeBody(resp, &summary); err != nil {
		return nil, err
	}

	var stats *PodStats
	for _, pod := range summary.Pods {
		if pod.PodRef.Namespace == c.namespace && pod.PodRef.Name == c.name {
			stats = &pod
			break
		}
	}
	if stats == nil {
		return nil, fmt.Errorf("pod %s.%s not found", c.namespace, c.name)
	}

	if c.ignoreContainer != nil {
		containers := make([]ContainerStats, 0, len(stats.Containers))
		for _, container := range stats.Containers {
			if c.ignoreContainer.MatchString(container.Name) {
				continue
			}
			containers = append(containers, container)
		}
		stats.Containers = containers
	}

	return stats, nil
}

// GetPodStats gets pod spec
func (c *client) GetSpec(ctx context.Context) (*cadvisorTypes.MachineInfo, error) {
	req, err := c.newRequest(specPath)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	var info cadvisorTypes.MachineInfo
	if err = decodeBody(resp, &info); err != nil {
		return nil, err
	}
	return &info, err
}

func (c *client) newRequest(endpoint string) (*http.Request, error) {
	u := *c.url
	u.Path = path.Join(c.url.Path, endpoint)
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	return req, nil
}

func decodeBody(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("got status code %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
