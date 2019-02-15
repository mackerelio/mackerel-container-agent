package instance

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// Client interface gets instance IPv4 address
type Client interface {
	GetLocalIPv4(context.Context) (string, error)
}

const (
	// DefaultURL is the default URL for EC2 instance metadata
	DefaultURL = "http://169.254.169.254"

	basePath      = "/latest/meta-data"
	localIPv4Path = "/local-ipv4"
)

var timeout = 3 * time.Second

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

// GetLocalIPv4 gets instance IPv4 address
func (c *client) GetLocalIPv4(ctx context.Context) (string, error) {
	req, err := c.newRequest(localIPv4Path)
	if err != nil {
		return "", err
	}
	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return "", err
	}
	body, err := readBody(resp)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(body)), nil
}

func (c *client) newRequest(endpoint string) (*http.Request, error) {
	u := *c.url
	u.Path = path.Join(c.url.Path, basePath, endpoint)
	return http.NewRequest("GET", u.String(), nil)
}

func readBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got status code %d", resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}
