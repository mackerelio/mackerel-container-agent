package spec

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mackerelio/golib/logging"
	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/api"
)

var logger = logging.GetLogger("spec")

// Manager in spec manager
type Manager struct {
	collector         *collector
	sender            *sender
	checks            []mackerel.CheckConfig
	version, revision string
	customIdentifier  string
}

// NewManager creates spec manager instanace
func NewManager(generators []Generator, client api.Client) *Manager {
	return &Manager{
		collector: newCollector(generators),
		sender:    newSender(client),
	}
}

// WithVersion sets agent version and revision
func (m *Manager) WithVersion(version, revision string) *Manager {
	m.version, m.revision = version, revision
	return m
}

// WithCustomIdentifier sets platform customIdentifier
func (m *Manager) WithCustomIdentifier(customIdentifier string) *Manager {
	m.customIdentifier = customIdentifier
	return m
}

// Run collect and send specs
func (m *Manager) Run(ctx context.Context, initialInterval, interval time.Duration) error {
	d := initialInterval
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(d):
			err := m.collectAndPost(ctx)
			if err != nil {
				// do not break the loop with spec posting error
				logger.Warningf("failed to update host spec: %s", err)
			}
			d = interval
		}
	}
	return nil
}

// Get collect specs
func (m *Manager) Get(ctx context.Context) (*mackerel.CreateHostParam, error) {
	var param mackerel.CreateHostParam
	name, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	meta, hostname, err := m.collector.collect(ctx)
	if err != nil {
		return nil, err
	}
	if hostname != "" {
		name = hostname
	}
	param.Name = name
	param.Meta = meta
	param.Meta.AgentName = BuildUserAgent(m.version, m.revision)
	param.Meta.AgentVersion = m.version + "-container"
	param.Meta.AgentRevision = m.revision
	ifaces, err := getInterfaces()
	if err != nil {
		return nil, err
	}
	param.Interfaces = ifaces
	param.CustomIdentifier = m.customIdentifier
	return &param, nil
}

// BuildUserAgent creates User-Agent, also used in agent-name of host's meta
func BuildUserAgent(version, revision string) string {
	return fmt.Sprintf("mackerel-container-agent/%s (Revision %s)", version, revision)
}

// SetHostID sets host id
func (m *Manager) SetHostID(hostID string) {
	m.sender.setHostID(hostID)
}

func (m *Manager) collectAndPost(ctx context.Context) error {
	param, err := m.Get(ctx)
	if err != nil {
		return err
	}
	updateParam := mackerel.UpdateHostParam(*param)
	updateParam.Checks = m.checks
	return m.sender.post(&updateParam)
}

// SetChecks sets check configs
func (m *Manager) SetChecks(checks []mackerel.CheckConfig) {
	m.checks = checks
}
