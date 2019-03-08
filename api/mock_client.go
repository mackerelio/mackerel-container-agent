package api

import mackerel "github.com/mackerelio/mackerel-client-go"

// MockClient represents a mock client of Mackerel API
type MockClient struct {
	findHostCallback                     func(id string) (*mackerel.Host, error)
	findHostsCallback                    func(param *mackerel.FindHostsParam) ([]*mackerel.Host, error)
	createHostCallback                   func(param *mackerel.CreateHostParam) (string, error)
	updateHostCallback                   func(hostID string, param *mackerel.UpdateHostParam) (string, error)
	updateHostStatusCallback             func(hostID string, status string) error
	retireHostCallback                   func(id string) error
	postHostMetricValuesByHostIDCallback func(hostID string, metricValues []*mackerel.MetricValue) error
	createGraphDefsCallback              func(graphDefs []*mackerel.GraphDefsParam) error
	postCheckReportsCallback             func(reports *mackerel.CheckReports) error
	metricValues                         map[string][]*mackerel.MetricValue
	graphDefs                            []*mackerel.GraphDefsParam
}

// MockClientOption represents an option of mock client of Mackerel API
type MockClientOption func(*MockClient)

// NewMockClient creates a new mock client of Mackerel API
func NewMockClient(opts ...MockClientOption) *MockClient {
	client := &MockClient{
		metricValues: make(map[string][]*mackerel.MetricValue),
	}
	for _, opt := range opts {
		client.ApplyOption(opt)
	}
	return client
}

// ApplyOption applies a mock client option
func (c *MockClient) ApplyOption(opt MockClientOption) {
	opt(c)
}

type errCallbackNotFound string

func (err errCallbackNotFound) Error() string {
	return string(err) + " callback not found"
}

// FindHost ...
func (c *MockClient) FindHost(id string) (*mackerel.Host, error) {
	if c.findHostCallback != nil {
		return c.findHostCallback(id)
	}
	return nil, errCallbackNotFound("FindHost")
}

// MockFindHost returns an option to set the callback of FindHost
func MockFindHost(callback func(id string) (*mackerel.Host, error)) MockClientOption {
	return func(c *MockClient) {
		c.findHostCallback = callback
	}
}

// FindHosts ...
func (c *MockClient) FindHosts(param *mackerel.FindHostsParam) ([]*mackerel.Host, error) {
	if c.findHostsCallback != nil {
		return c.findHostsCallback(param)
	}
	return nil, errCallbackNotFound("FindHosts")
}

// MockFindHosts returns an option to set the callback of FindHosts
func MockFindHosts(callback func(param *mackerel.FindHostsParam) ([]*mackerel.Host, error)) MockClientOption {
	return func(c *MockClient) {
		c.findHostsCallback = callback
	}
}

// CreateHost ...
func (c *MockClient) CreateHost(param *mackerel.CreateHostParam) (string, error) {
	if c.createHostCallback != nil {
		return c.createHostCallback(param)
	}
	return "", errCallbackNotFound("CreateHost")
}

// MockCreateHost returns an option to set the callback of CreateHost
func MockCreateHost(callback func(param *mackerel.CreateHostParam) (string, error)) MockClientOption {
	return func(c *MockClient) {
		c.createHostCallback = callback
	}
}

// UpdateHost ...
func (c *MockClient) UpdateHost(hostID string, param *mackerel.UpdateHostParam) (string, error) {
	if c.updateHostCallback != nil {
		return c.updateHostCallback(hostID, param)
	}
	return "", errCallbackNotFound("UpdateHost")
}

// MockUpdateHost returns an option to set the callback of UpdateHost
func MockUpdateHost(callback func(hostID string, param *mackerel.UpdateHostParam) (string, error)) MockClientOption {
	return func(c *MockClient) {
		c.updateHostCallback = callback
	}
}

// UpdateHostStatus ...
func (c *MockClient) UpdateHostStatus(hostID string, status string) error {
	if c.updateHostStatusCallback != nil {
		return c.updateHostStatusCallback(hostID, status)
	}
	return errCallbackNotFound("UpdateHostStatus")
}

// MockUpdateHostStatus returns an option to set the callback of UpdateHostStatus
func MockUpdateHostStatus(callback func(hostID string, status string) error) MockClientOption {
	return func(c *MockClient) {
		c.updateHostStatusCallback = callback
	}
}

// RetireHost ...
func (c *MockClient) RetireHost(id string) error {
	if c.retireHostCallback != nil {
		return c.retireHostCallback(id)
	}
	return errCallbackNotFound("RetireHost")
}

// MockRetireHost returns an option to set the callback of RetireHost
func MockRetireHost(callback func(id string) error) MockClientOption {
	return func(c *MockClient) {
		c.retireHostCallback = callback
	}
}

// PostHostMetricValuesByHostID ...
func (c *MockClient) PostHostMetricValuesByHostID(hostID string, metricValues []*mackerel.MetricValue) error {
	if c.postHostMetricValuesByHostIDCallback != nil {
		return c.postHostMetricValuesByHostIDCallback(hostID, metricValues)
	}
	if _, ok := c.metricValues[hostID]; ok {
		c.metricValues[hostID] = append(c.metricValues[hostID], metricValues...)
	} else {
		c.metricValues[hostID] = metricValues
	}
	return nil
}

// MockPostHostMetricValuesByHostID returns an option to set the callback of PostHostMetricValuesByHostID
func MockPostHostMetricValuesByHostID(callback func(hostID string, metricValues []*mackerel.MetricValue) error) MockClientOption {
	return func(c *MockClient) {
		c.postHostMetricValuesByHostIDCallback = callback
	}
}

// CreateGraphDefs ...
func (c *MockClient) CreateGraphDefs(graphDefs []*mackerel.GraphDefsParam) error {
	if c.createGraphDefsCallback != nil {
		return c.createGraphDefsCallback(graphDefs)
	}
	c.graphDefs = append(c.graphDefs, graphDefs...)
	return nil
}

// MockCreateGraphDefs returns an option to set the callback of CreateGraphDefs
func MockCreateGraphDefs(callback func(graphDefs []*mackerel.GraphDefsParam) error) MockClientOption {
	return func(c *MockClient) {
		c.createGraphDefsCallback = callback
	}
}

// PostCheckReports ...
func (c *MockClient) PostCheckReports(reports *mackerel.CheckReports) error {
	if c.postCheckReportsCallback != nil {
		return c.postCheckReportsCallback(reports)
	}
	return nil
}

// MockPostCheckReports returns an option to set the callback of PostCheckReports
func MockPostCheckReports(callback func(reports *mackerel.CheckReports) error) MockClientOption {
	return func(c *MockClient) {
		c.postCheckReportsCallback = callback
	}
}

// PostedMetricValues returns the posted metric values
func (c *MockClient) PostedMetricValues() map[string][]*mackerel.MetricValue {
	return c.metricValues
}

// PostedGraphDefs returns the posted graph definitions
func (c *MockClient) PostedGraphDefs() []*mackerel.GraphDefsParam {
	return c.graphDefs
}
