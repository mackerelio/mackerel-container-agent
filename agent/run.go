package agent

import (
	"context"
	"time"

	"github.com/Songmu/retry"
	"golang.org/x/sync/errgroup"

	"github.com/mackerelio/mackerel-container-agent/api"
	"github.com/mackerelio/mackerel-container-agent/check"
	"github.com/mackerelio/mackerel-container-agent/config"
	"github.com/mackerelio/mackerel-container-agent/metric"
	"github.com/mackerelio/mackerel-container-agent/platform"
	"github.com/mackerelio/mackerel-container-agent/probe"
	"github.com/mackerelio/mackerel-container-agent/spec"
)

var (
	metricsInterval            = time.Minute
	checkInterval              = time.Minute
	specInterval               = time.Hour
	specInitialInterval        = 5 * time.Minute
	waitStatusRunningInterval  = 3 * time.Second
	hostIDInitialRetryInterval = 1 * time.Second
)

func run(
	ctx context.Context,
	client api.Client,
	metricManager *metric.Manager,
	checkManager *check.Manager,
	specManager *spec.Manager,
	pform platform.Platform,
	conf *config.Config,
) (func(), error) {
	specManager.SetChecks(checkManager.Configs())
	eg, ctx := errgroup.WithContext(ctx)

	hostResolver := newHostResolver(client, conf.Root)
	eg.Go(func() error {
		var duration time.Duration
	loop:
		for {
			select {
			case <-time.After(duration):
				if pform.StatusRunning(ctx) {
					break loop
				}
				if duration == 0 {
					duration = waitStatusRunningInterval
				}
				logger.Infof("wait for the platform status to be running")
			case <-ctx.Done():
				return nil
			}
		}

		if conf.ReadinessProbe != nil {
			if err := probe.Wait(ctx, probe.NewProbe(conf.ReadinessProbe)); err != nil {
				return nil
			}
		}

		hostParam, err := specManager.Get(ctx)
		if err != nil {
			return err
		}
		hostParam.RoleFullnames = conf.Roles
		hostParam.DisplayName = conf.DisplayName
		hostParam.Memo = conf.Memo
		hostParam.Checks = checkManager.Configs()

		duration = hostIDInitialRetryInterval
		for {
			select {
			case <-time.After(duration):
				host, retryHostID, err := hostResolver.getHost(hostParam)
				if retryHostID {
					logger.Infof("retry to find host: %s", err)
					if duration *= 2; duration > 10*time.Minute {
						duration = 10 * time.Minute
					}
					continue
				}
				if err != nil {
					return err
				}
				logger.Infof("start the agent: host id = %s, host name = %s", host.ID, hostParam.Name)
				if conf.HostStatusOnStart != "" && host.Status != string(conf.HostStatusOnStart) {
					err = retry.Retry(5, 3*time.Second, func() error {
						return client.UpdateHostStatus(host.ID, string(conf.HostStatusOnStart))
					})
					if err != nil {
						logger.Warningf("failed to update host status on start: %s", err)
					}
				}
				err = retry.Retry(5, 3*time.Second, func() error {
					return metricManager.CollectAndPostGraphDefs(ctx)
				})
				if err != nil {
					logger.Warningf("failed to post graph definitions: %s", err)
				}
				metricManager.SetHostID(host.ID)
				checkManager.SetHostID(host.ID)
				specManager.SetHostID(host.ID)
				return nil
			case <-ctx.Done():
				return nil
			}
		}
	})

	eg.Go(func() error {
		return metricManager.Run(ctx, metricsInterval)
	})

	eg.Go(func() error {
		return checkManager.Run(ctx, checkInterval)
	})

	eg.Go(func() error {
		return specManager.Run(ctx, specInitialInterval, specInterval)
	})

	return func() {
		if err := retire(client, hostResolver); err != nil {
			logger.Warningf("failed to retire: %s", err)
		}
	}, eg.Wait()
}
