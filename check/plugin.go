package check

import (
	"context"
	"time"

	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/cmdutil"
	"github.com/mackerelio/mackerel-container-agent/config"
)

type pluginGenerator struct {
	config.CheckPlugin
	lastResult *Result
}

// NewPluginGenerator creates a new check generator
func NewPluginGenerator(p *config.CheckPlugin) Generator {
	return &pluginGenerator{*p, nil}
}

// Config gets check generator config
func (g *pluginGenerator) Config() mackerel.CheckConfig {
	return mackerel.CheckConfig{Name: g.Name, Memo: g.Memo}
}

// Generate generates check report
func (g *pluginGenerator) Generate(ctx context.Context) (*Result, error) {
	now := time.Now()
	stdout, stderr, exitCode, err := cmdutil.RunCommand(ctx, g.Command, g.User, g.Env, g.Timeout)

	if stderr != "" {
		logger.Infof("plugin %s (%s): %q", g.Name, g.Command, stderr)
	}

	var message string
	var status mackerel.CheckStatus
	if err != nil {
		logger.Warningf("plugin %s (%s): %s", g.Name, g.Command, err)
		message = err.Error()
		status = mackerel.CheckStatusUnknown
	} else {
		message = stdout
		status = exitCodeToStatus(exitCode)
	}

	newResult := NewResult(g.Name, message, status, now)

	lastResult := g.lastResult
	g.lastResult = newResult
	if lastResult == nil {
		return newResult, nil
	}
	if lastResult.status == mackerel.CheckStatusOK && newResult.status == mackerel.CheckStatusOK {
		// do not report ok -> ok
		return nil, nil
	}
	return newResult, nil
}

func exitCodeToStatus(exitCode int) mackerel.CheckStatus {
	switch exitCode {
	case 0:
		return mackerel.CheckStatusOK
	case 1:
		return mackerel.CheckStatusWarning
	case 2:
		return mackerel.CheckStatusCritical
	default:
		return mackerel.CheckStatusUnknown
	}
}
