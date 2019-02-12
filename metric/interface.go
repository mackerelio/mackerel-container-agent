package metric

import (
	"context"
	"strings"
	"time"

	"github.com/mackerelio/go-osstat/network"
	mackerel "github.com/mackerelio/mackerel-client-go"
)

type interfaceGenerator struct {
	prevStats map[string]network.Stats
	prevTime  time.Time
}

// NewInterfaceGenerator creates interface generator
func NewInterfaceGenerator() Generator {
	return &interfaceGenerator{}
}

func (g *interfaceGenerator) Generate(context.Context) (Values, error) {
	stats, err := g.getInterfaceStats()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if g.prevStats == nil || g.prevTime.Before(now.Add(-10*time.Minute)) {
		g.prevStats = stats
		g.prevTime = now
		return nil, nil
	}

	values := make(Values)
	timeDelta := now.Sub(g.prevTime).Seconds()
	for name, prevValue := range g.prevStats {
		currValue, ok := stats[name]
		if !ok {
			continue
		}
		name = SanitizeMetricKey(name)
		prefix := "interface." + name
		values[prefix+".rxBytes.delta"] = float64(currValue.RxBytes-prevValue.RxBytes) / timeDelta
		values[prefix+".txBytes.delta"] = float64(currValue.TxBytes-prevValue.TxBytes) / timeDelta
	}

	g.prevStats = stats
	g.prevTime = now

	return values, nil
}

func (g *interfaceGenerator) getInterfaceStats() (map[string]network.Stats, error) {
	stats, err := network.Get()
	if err != nil {
		return nil, err
	}
	values := make(map[string]network.Stats)
	for _, s := range stats {
		if strings.HasPrefix(s.Name, "veth") {
			continue
		}
		values[s.Name] = s
	}
	return values, nil
}

func (g *interfaceGenerator) GetGraphDefs(context.Context) ([]*mackerel.GraphDefsParam, error) {
	return nil, nil
}
