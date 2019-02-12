package taskmetadata

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"time"

	ecsTypes "github.com/aws/amazon-ecs-agent/agent/handlers/v2"
	dockerTypes "github.com/docker/docker/api/types"
)

// Client interface gets metadata and stats
type Client interface {
	GetMetadata(context.Context) (*ecsTypes.TaskResponse, error)
	GetStats(context.Context) (map[string]*dockerTypes.Stats, error)
}

const (
	// DefaultURL represents Task Metadata URL
	DefaultURL = "http://169.254.170.2"

	metadataPath = "/v2/metadata"
	statsPath    = "/v2/stats"
)

var timeout = 3 * time.Second

type client struct {
	url             *url.URL
	httpClient      *http.Client
	ignoreContainer *regexp.Regexp
}

// NewClient creates a new Client
func NewClient(baseURL string, ignoreContainer *regexp.Regexp) (Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	return &client{
		url: u,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		ignoreContainer: ignoreContainer,
	}, nil
}

// GetMetadata gets task metadata
func (c *client) GetMetadata(ctx context.Context) (*ecsTypes.TaskResponse, error) {
	req, err := c.newRequest(metadataPath)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	var data ecsTypes.TaskResponse
	if err = decodeBody(resp, &data); err != nil {
		return nil, err
	}
	if c.ignoreContainer != nil {
		containers := make([]ecsTypes.ContainerResponse, 0, len(data.Containers))
		for _, container := range data.Containers {
			if c.ignoreContainer.MatchString(container.Name) {
				continue
			}
			containers = append(containers, container)
		}
		data.Containers = containers
	}
	return &data, nil
}

// GetMetadata gets task stats
func (c *client) GetStats(ctx context.Context) (map[string]*dockerTypes.Stats, error) {
	req, err := c.newRequest(statsPath)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	var data map[string]*dockerTypes.Stats
	if err = decodeBody(resp, &data); err != nil {
		return nil, err
	}

	meta, err := c.GetMetadata(ctx)
	if err != nil {
		return nil, err
	}

	stats := make(map[string]*dockerTypes.Stats)

	for _, container := range meta.Containers {
		if v, ok := data[container.ID]; ok {
			stats[container.ID] = v
		}
	}

	return stats, nil
}

func (c *client) newRequest(endpoint string) (*http.Request, error) {
	u := *c.url
	u.Path = path.Join(c.url.Path, endpoint)
	return http.NewRequest("GET", u.String(), nil)
}

func decodeBody(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}
