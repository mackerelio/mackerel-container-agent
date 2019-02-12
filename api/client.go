package api

import mackerel "github.com/mackerelio/mackerel-client-go"

// Client represents a client of Mackerel API
type Client interface {
	FindHost(id string) (*mackerel.Host, error)
	FindHosts(param *mackerel.FindHostsParam) ([]*mackerel.Host, error)
	CreateHost(param *mackerel.CreateHostParam) (string, error)
	UpdateHost(hostID string, param *mackerel.UpdateHostParam) (string, error)
	UpdateHostStatus(hostID string, status string) error
	RetireHost(id string) error
	PostHostMetricValuesByHostID(hostID string, metricValues []*mackerel.MetricValue) error
	CreateGraphDefs([]*mackerel.GraphDefsParam) error
	PostCheckReports(reports *mackerel.CheckReports) error
}
