package probe

import (
	"context"
	"time"

	"github.com/mackerelio/golib/logging"

	"github.com/mackerelio/mackerel-container-agent/config"
)

var logger = logging.GetLogger("probe")

var (
	defaultPeriod = 10 * time.Second
)

// Probe ...
type Probe interface {
	Check(context.Context) error
	InitialDelay() time.Duration
	Period() time.Duration
}

// NewProbe creates a new Probe
func NewProbe(p *config.Probe) Probe {
	initialDelay := time.Duration(p.InitialDelaySeconds) * time.Second
	period := time.Duration(p.PeriodSeconds) * time.Second
	timeout := time.Duration(p.TimeoutSeconds) * time.Second
	if p.Exec != nil {
		return &probeExec{
			ProbeExec:    p.Exec,
			initialDelay: initialDelay,
			period:       period,
			timeout:      timeout,
		}
	}
	if p.HTTP != nil {
		return &probeHTTP{
			ProbeHTTP:    p.HTTP,
			initialDelay: initialDelay,
			period:       period,
			timeout:      timeout,
		}
	}
	if p.TCP != nil {
		return &probeTCP{
			ProbeTCP:     p.TCP,
			initialDelay: initialDelay,
			period:       period,
			timeout:      timeout,
		}
	}
	return nil
}

// Wait until the probe is ready.
func Wait(ctx context.Context, p Probe) error {
	if delay := p.InitialDelay(); delay > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	period := p.Period()
	if period == 0 {
		period = defaultPeriod
	}
loop:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err := p.Check(ctx)
			if err != nil {
				logger.Infof("%s", err)
			} else {
				break loop
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(period):
		}
	}
	return nil
}
