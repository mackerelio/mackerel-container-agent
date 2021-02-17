package agent

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/mackerelio/golib/logging"
	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/check"
	"github.com/mackerelio/mackerel-container-agent/config"
	"github.com/mackerelio/mackerel-container-agent/metric"
	"github.com/mackerelio/mackerel-container-agent/spec"
)

var logger = logging.GetLogger("agent")

// Agent interface
type Agent interface {
	Run([]string) error
}

// NewAgent creates a new Mackerel agent
func NewAgent(version, revision string) Agent {
	return &agent{version, revision}
}

type agent struct {
	version, revision string
}

func (a *agent) Run(_ []string) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP)
	retires := make([]func(), 0, 1)
	defer func() {
		for _, retire := range retires {
			retire()
		}
	}()
	confLoader, err := createConfLoader()
	if err != nil {
		return err
	}
	for {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		conf, err := confLoader.Load(ctx)
		if err != nil {
			return err
		}
		errCh := make(chan error)
		go func() {
			retire, err := a.start(ctx, conf)
			if retire != nil {
				retires = append(retires, retire)
			}
			errCh <- err
		}()
		confCh := confLoader.Start(ctx)
		select {
		case sig := <-sigCh:
			logger.Infof("reload config: signal = %s", sig)
			cancel()
		case <-confCh:
			cancel()
		case err := <-errCh:
			return err
		}
	}
}

func createConfLoader() (*config.Loader, error) {
	var pollingDuration time.Duration
	if durationMinutesStr := os.Getenv(
		"MACKEREL_AGENT_CONFIG_POLLING_DURATION_MINUTES",
	); durationMinutesStr != "" {
		durationMinutes, err := strconv.Atoi(durationMinutesStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse config polling duration: %w", err)
		}
		pollingDuration = time.Duration(durationMinutes) * time.Minute
	}
	return config.NewLoader(os.Getenv("MACKEREL_AGENT_CONFIG"), pollingDuration), nil
}

func (a *agent) start(ctx context.Context, conf *config.Config) (func(), error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	client := mackerel.NewClient(conf.Apikey)
	if conf.Apibase != "" {
		baseURL, err := url.Parse(conf.Apibase)
		if err != nil {
			return nil, err
		}
		client.BaseURL = baseURL
	}
	client.UserAgent = spec.BuildUserAgent(a.version, a.revision)
	if conf.ReadinessProbe != nil && conf.ReadinessProbe.HTTP != nil {
		conf.ReadinessProbe.HTTP.UserAgent = client.UserAgent
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer signal.Stop(sigCh)
	var sig os.Signal
	go func() {
		select {
		case sig = <-sigCh:
			cancel()
		case <-ctx.Done():
		}
	}()
	defer func() {
		if sig != nil {
			logger.Infof("stop the agent: signal = %s", sig)
		}
	}()

	pform, err := NewPlatform(ctx, conf.IgnoreContainer.Regexp)
	if err != nil {
		return nil, err
	}

	customIdentifier, err := pform.GetCustomIdentifier(ctx)
	if err != nil {
		logger.Warningf("failed to get custom identifier: %s", err)
	}

	metricGenerators := pform.GetMetricGenerators()
	for _, mp := range conf.MetricPlugins {
		metricGenerators = append(metricGenerators, metric.NewPluginGenerator(mp))
	}
	metricManager := metric.NewManager(metricGenerators, client)

	var checkGenerators []check.Generator
	for _, cp := range conf.CheckPlugins {
		checkGenerators = append(checkGenerators, check.NewPluginGenerator(cp))
	}
	checkManager := check.NewManager(checkGenerators, client)

	specGenerators := pform.GetSpecGenerators()
	specManager := spec.NewManager(specGenerators, client).
		WithVersion(a.version, a.revision).
		WithCustomIdentifier(customIdentifier)

	return run(ctx, client, metricManager, checkManager, specManager, pform, conf)
}
