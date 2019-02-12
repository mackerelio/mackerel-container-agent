package probe

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/mackerelio/mackerel-container-agent/config"
)

var (
	defaultTimeoutTCP = 1 * time.Second
)

type probeTCP struct {
	*config.ProbeTCP
	initialDelay time.Duration
	period       time.Duration
	timeout      time.Duration
}

func (p *probeTCP) Check(ctx context.Context) error {
	timeout := p.timeout
	if timeout == 0 {
		timeout = defaultTimeoutTCP
	}

	addr := net.JoinHostPort(p.Host, p.Port)
	d := net.Dialer{Timeout: timeout}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("tcp probe failed (%s): %s", addr, err)
	}
	defer conn.Close()

	logger.Infof("tcp probe success (%s)", addr)
	return nil
}

func (p *probeTCP) InitialDelay() time.Duration {
	return p.initialDelay
}

func (p *probeTCP) Period() time.Duration {
	return p.period
}
