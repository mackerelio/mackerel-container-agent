package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	ecsTypes "github.com/aws/amazon-ecs-agent/agent/handlers/v1"
)

const (
	// DefaultPort represents ECS agent introspection API port number
	DefaultPort = 51678

	metadataPath = "/v1/metadata"
	taskPath     = "/v1/tasks"
)

var timeout = 3 * time.Second

// Client interface gets metadata
type Client interface {
	GetInstanceMetadata(context.Context) (*ecsTypes.MetadataResponse, error)
	GetTaskMetadataWithDockerID(context.Context, string) (*ecsTypes.TaskResponse, error)
}

type client struct {
	url        *url.URL
	httpClient *http.Client
}

// NewClient creates a new Client
func NewClient(baseURL string) (Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	return &client{
		url: u,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// GetInstanceMetadata gets instance metadata
func (c *client) GetInstanceMetadata(ctx context.Context) (*ecsTypes.MetadataResponse, error) {
	req, err := c.newRequest(metadataPath)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	var data ecsTypes.MetadataResponse
	if err := decodeBody(resp, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// GetTaskMetadataWithDockerID gets task metadata
func (c *client) GetTaskMetadataWithDockerID(ctx context.Context, dockerID string) (*ecsTypes.TaskResponse, error) {
	req, err := c.newRequest(taskPath)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Set("dockerid", dockerID)
	req.URL.RawQuery = q.Encode()
	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	var data ecsTypes.TaskResponse
	if err := decodeBody(resp, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (c *client) newRequest(endpoint string) (*http.Request, error) {
	u := *c.url
	u.Path = path.Join(c.url.Path, endpoint)
	return http.NewRequest("GET", u.String(), nil)
}

func decodeBody(resp *http.Response, out interface{}) error {
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("incorrect status code %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
