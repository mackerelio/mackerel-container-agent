package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path"
	"time"

	dockerTypes "github.com/docker/docker/api/types"
)

const (
	// DefaultSock represents path to docker.sock
	DefaultSock = "/var/run/docker.sock"

	statsPath   = "/containers/%s/stats"
	inspectPath = "/containers/%s/json"
)

var timeout = 3 * time.Second

// Client interface gets container stats and specs
type Client interface {
	GetContainerStats(context.Context, string) (*dockerTypes.StatsJSON, error)
	InspectContainer(context.Context, string) (*dockerTypes.ContainerJSON, error)
}

type client struct {
	url        *url.URL
	httpClient *http.Client
}

// NewClient creates a new Client
func NewClient(sock string) (Client, error) {
	u, _ := url.Parse("http://localhost")
	return &client{
		url: u,
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					return net.Dial("unix", sock)
				},
			},
		},
	}, nil
}

// GetContainerStats gets container stats
func (c *client) GetContainerStats(ctx context.Context, id string) (*dockerTypes.StatsJSON, error) {
	req, err := c.newRequest(fmt.Sprintf(statsPath, id))
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Set("stream", "false")
	req.URL.RawQuery = q.Encode()
	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	var data dockerTypes.StatsJSON
	if err := decodeBody(resp, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// InspectContainer gets container specs
func (c *client) InspectContainer(ctx context.Context, id string) (*dockerTypes.ContainerJSON, error) {
	req, err := c.newRequest(fmt.Sprintf(inspectPath, id))
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	var data dockerTypes.ContainerJSON
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
