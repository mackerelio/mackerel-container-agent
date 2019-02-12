package probe

import (
	"context"
	"fmt"
	"time"

	"github.com/mackerelio/mackerel-container-agent/cmdutil"
	"github.com/mackerelio/mackerel-container-agent/config"
)

var (
	defaultTimeoutExec = 1 * time.Second
)

type probeExec struct {
	*config.ProbeExec
	initialDelay time.Duration
	period       time.Duration
	timeout      time.Duration
}

func (p *probeExec) Check(ctx context.Context) error {
	timeout := p.timeout
	if timeout == 0 {
		timeout = defaultTimeoutExec
	}
	_, stderr, exitCode, err := cmdutil.RunCommand(ctx, p.Command, p.User, p.Env, timeout)

	if stderr != "" {
		stderr = fmt.Sprintf(", stderr = %q", stderr)
	}
	if err != nil {
		return fmt.Errorf("exec probe failed (%s): %s%s", p.Command, err, stderr)
	}
	if exitCode != 0 {
		return fmt.Errorf("exec probe failed (%s): exit code = %d%s", p.Command, exitCode, stderr)
	}

	logger.Infof("exec probe success (%s): exit code = %d", p.Command, exitCode)
	return nil
}

func (p *probeExec) InitialDelay() time.Duration {
	return p.initialDelay
}

func (p *probeExec) Period() time.Duration {
	return p.period
}
