package kubernetesapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	kubernetesTypes "k8s.io/api/core/v1"
)

// Client interface gets metadata and stats
type Client interface {
	GetNode(context.Context) (*kubernetesTypes.Node, error)
}

const (
	basePath = "/api/v1"
	nodePath = "/nodes/"
)

type client struct {
	url        *url.URL
	httpClient *http.Client
	nodeName   string
	token      string
}

// NewClient creates a new Client
func NewClient(httpClient *http.Client, token, baseURL, nodeName string) (Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	return &client{
		url:        u,
		httpClient: httpClient,
		token:      token,
		nodeName:   nodeName,
	}, nil
}

// GetNode gets node spec
func (c *client) GetNode(ctx context.Context) (*kubernetesTypes.Node, error) {
	req, err := c.newRequest(nodePath + c.nodeName)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	var node kubernetesTypes.Node
	if err = decodeBody(resp, &node); err != nil {
		return nil, err
	}
	return &node, err
}

func (c *client) newRequest(endpoint string) (*http.Request, error) {
	u := *c.url
	u.Path = path.Join(c.url.Path, basePath, endpoint)
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
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("got status code %d (url: %s, body: %q)", resp.StatusCode, resp.Request.URL, body)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
