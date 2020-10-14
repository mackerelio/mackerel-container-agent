package kubernetesapi

import (
	"context"
	"net/url"

	kubernetesTypes "k8s.io/api/core/v1"
)

// MockClient represents a mock client of Kubernetes APIs
type MockClient struct {
	getNodeCallback      func(context.Context) (*kubernetesTypes.Node, error)
	nodeProxyURLCallback func() *url.URL
}

// MockClientOption represents an option of mock client of Kubernetes APIs
type MockClientOption func(*MockClient)

// NewMockClient creates a new mock client of Kubernetes APIs
func NewMockClient(opts ...MockClientOption) *MockClient {
	c := &MockClient{}
	for _, o := range opts {
		c.ApplyOption(o)
	}
	return c
}

// ApplyOption applies a mock client option
func (c *MockClient) ApplyOption(opt MockClientOption) {
	opt(c)
}

type errCallbackNotFound string

func (err errCallbackNotFound) Error() string {
	return string(err) + " callback not found"
}

// GetNode ...
func (c *MockClient) GetNode(ctx context.Context) (*kubernetesTypes.Node, error) {
	if c.getNodeCallback != nil {
		return c.getNodeCallback(ctx)
	}
	return nil, errCallbackNotFound("GetNode")
}

// MockGetNode returns an option to set the callback of GetNode
func MockGetNode(callback func(context.Context) (*kubernetesTypes.Node, error)) MockClientOption {
	return func(c *MockClient) {
		c.getNodeCallback = callback
	}
}

// NodeProxyURL ...
func (c *MockClient) NodeProxyURL() *url.URL {
	if c.nodeProxyURLCallback != nil {
		return c.nodeProxyURLCallback()
	}
	return nil
}

// MockNodeProxyURL returns an option to set the callback of NodeProxyURL
func MockNodeProxyURL(callback func() *url.URL) MockClientOption {
	return func(c *MockClient) {
		c.nodeProxyURLCallback = callback
	}
}
