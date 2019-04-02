package taskmetadata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"time"

	ecsTypes "github.com/aws/amazon-ecs-agent/agent/handlers/v2"
)

const (
	metadataPath = "/task"
	statsPath    = "/task/stats"
)

var timeout = 3 * time.Second

// Client ...
type Client struct {
	url             *url.URL
	httpClient      *http.Client
	ignoreContainer *regexp.Regexp
}

// NewClient creates a new Client
func NewClient(baseURL string, ignoreContainer *regexp.Regexp) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	return &Client{
		url: u,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		ignoreContainer: ignoreContainer,
	}, nil
}

// GetTaskMetadata gets task metadata
func (c *Client) GetTaskMetadata(ctx context.Context) (*ecsTypes.TaskResponse, error) {
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

// // GetTaskStats gets task stats
// func (c *Client) GetTaskStats(ctx context.Context) (map[string]*dockerTypes.Stats, error) {
//   req, err := c.newRequest(statsPath)
//   if err != nil {
//     return nil, err
//   }
//   resp, err := c.httpClient.Do(req.WithContext(ctx))
//   if err != nil {
//     return nil, err
//   }
//   var data map[string]*dockerTypes.Stats
//   if err = decodeBody(resp, &data); err != nil {
//     return nil, err
//   }

//   meta, err := c.GetTaskMetadata(ctx)
//   if err != nil {
//     return nil, err
//   }

//   stats := make(map[string]*dockerTypes.Stats)

//   for _, container := range meta.Containers {
//     if v, ok := data[container.ID]; ok {
//       stats[container.ID] = v
//     }
//   }

//   return stats, nil
// }

func (c *Client) newRequest(endpoint string) (*http.Request, error) {
	u := *c.url
	u.Path = path.Join(c.url.Path, endpoint)
	return http.NewRequest("GET", u.String(), nil)
}

func decodeBody(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("got status code %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
