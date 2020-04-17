package metric

import (
	"context"
	"time"

	"github.com/mackerelio/golib/logging"
	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/api"
)

var logger = logging.GetLogger("metric")

// Manager in metric manager
type Manager struct {
	collector *collector
	sender    *sender
}

// NewManager creates metric manager instanace
func NewManager(generators []Generator, client api.Client) *Manager {
	return &Manager{
		collector: newCollector(generators),
		sender:    newSender(client),
	}
}

// Run collect and send metrics
func (m *Manager) Run(ctx context.Context, interval time.Duration) (err error) {
	t := time.NewTicker(interval)
	defer t.Stop()
	errCh := make(chan error)
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-t.C:
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
				defer cancel()
				if err := m.collectAndPostValues(ctx); err != nil {
					errCh <- err
				}
			}()
		case err = <-errCh:
			break loop
		}
	}
	return
}

// SetHostID sets host id
func (m *Manager) SetHostID(hostID string) {
	m.sender.setHostID(hostID)
}

func (m *Manager) collectAndPostValues(ctx context.Context) error {
	now := time.Now()
	values, err := m.collector.collect(ctx)
	if err != nil {
		return err
	}
	if len(values) == 0 {
		return nil
	}
	var metricValues []*mackerel.MetricValue
	for name, value := range values {
		metricValues = append(metricValues, &mackerel.MetricValue{
			Name:  name,
			Time:  now.Unix(),
			Value: value,
		})
	}
	return m.sender.post(metricValues)
}

// CollectAndPostGraphDefs sends graph definitions
func (m *Manager) CollectAndPostGraphDefs(ctx context.Context) error {
	graphDefs, err := m.collector.collectGraphDefs(ctx)
	if err != nil {
		return err
	}
	return m.sender.postGraphDefs(graphDefs)
}
