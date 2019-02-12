package agent

import (
	"context"
	"net/url"
	"os"
	"os/signal"
	"syscall"

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
	for {
		errCh := make(chan error)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() { errCh <- a.start(ctx) }()
		select {
		case <-sigCh:
			cancel()
		case err := <-errCh:
			return err
		}
	}
}

func (a *agent) start(ctx context.Context) error {
	conf, err := config.Load(os.Getenv("MACKEREL_AGENT_CONFIG"))
	if err != nil {
		return err
	}

	client := mackerel.NewClient(conf.Apikey)
	if conf.Apibase != "" {
		baseURL, err := url.Parse(conf.Apibase)
		if err != nil {
			return err
		}
		client.BaseURL = baseURL
	}
	client.UserAgent = spec.BuildUserAgent(a.version, a.revision)
	if conf.ReadinessProbe != nil && conf.ReadinessProbe.HTTP != nil {
		conf.ReadinessProbe.HTTP.UserAgent = client.UserAgent
	}

	pform, err := NewPlatform(ctx, conf.IgnoreContainer.Regexp)
	if err != nil {
		return err
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

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer signal.Stop(sigCh)

	return run(ctx, client, metricManager, checkManager, specManager, pform, conf, sigCh)
}
