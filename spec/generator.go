package spec

import (
	"context"

	mackerel "github.com/mackerelio/mackerel-client-go"
)

// Generator interface generates spec information
type Generator interface {
	Generate(context.Context) (any, error)
}

// CloudHostname has mackerel.Cloud and host name
type CloudHostname struct {
	Cloud    *mackerel.Cloud
	Hostname string
}
