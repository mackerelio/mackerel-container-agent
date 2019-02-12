package check

import (
	"context"
	"time"

	"github.com/mackerelio/golib/logging"
	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/api"
)

var logger = logging.GetLogger("check")

// Manager represents check manager
type Manager struct {
	collector *collector
	sender    *sender
}

// NewManager creates a new check manager
func NewManager(generators []Generator, client api.Client) *Manager {
	return &Manager{
		collector: newCollector(generators),
		sender:    newSender(client),
	}
}

// Configs gets check manager configs
func (m *Manager) Configs() []mackerel.CheckConfig {
	return m.collector.configs()
}

// Run collect and check monitoring reports
func (m *Manager) Run(ctx context.Context, interval time.Duration) (err error) {
	t := time.NewTicker(interval)
	errCh := make(chan error)
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-t.C:
			go func() {
				if err := m.collectAndPostCheckReports(ctx); err != nil {
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

func (m *Manager) collectAndPostCheckReports(ctx context.Context) error {
	rs := m.collector.collect(ctx)
	reports := make([]*mackerel.CheckReport, len(rs))
	for i, r := range rs {
		reports[i] = &mackerel.CheckReport{
			Name:       r.name,
			Status:     r.status,
			Message:    r.message,
			OccurredAt: r.occurredAt.Unix(),
		}
	}
	return m.sender.post(reports)
}
