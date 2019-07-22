package spec

import (
	"context"

	mackerel "github.com/mackerelio/mackerel-client-go"
)

type collector struct {
	generators []Generator
}

func newCollector(generators []Generator) *collector {
	return &collector{
		generators: generators,
	}
}

func (c *collector) collect(ctx context.Context) (mackerel.HostMeta, string, error) {
	var ret mackerel.HostMeta
	var hostname string

	for _, g := range c.generators {
		v, err := g.Generate(ctx)
		if err != nil {
			return ret, hostname, err
		}
		switch v := v.(type) {
		case *mackerel.Cloud:
			ret.Cloud = v
		case *CloudHostname:
			ret.Cloud = v.Cloud
			hostname = v.Hostname
		case mackerel.CPU:
			ret.CPU = v
		default:
		}
	}

	return ret, hostname, nil
}
