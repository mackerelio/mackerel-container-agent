package metric

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/cmdutil"
	"github.com/mackerelio/mackerel-container-agent/config"
)

const (
	pluginPrefix       = "custom."
	pluginMetaEnvName  = "MACKEREL_AGENT_PLUGIN_META"
	pluginMetaHeadline = "# mackerel-agent-plugin"
)

type pluginGenerator struct {
	config.MetricPlugin
}

// NewPluginGenerator creates a new plugin generator
func NewPluginGenerator(p *config.MetricPlugin) Generator {
	return &pluginGenerator{*p}
}

// Generate generates metric values
func (g *pluginGenerator) Generate(ctx context.Context) (Values, error) {
	env := append(g.Env, pluginMetaEnvName+"=")
	var masked_env []string
	for _, v := range env {
		key := strings.Split(v, "=")[0]
		value := strings.Split(v, "=")[1]
		masked_env = append(masked_env, key+"="+config.MaskEnvValue(value))
	}
	logger.Debugf("plugin %s command: %s env: %+v", g.Name, g.Command, masked_env)
	stdout, stderr, _, err := cmdutil.RunCommand(ctx, g.Command, g.User, env, g.Timeout)

	if stderr != "" {
		logger.Infof("plugin %s (%s): %q", g.Name, g.Command, stderr)
	}
	if err != nil {
		return nil, fmt.Errorf("plugin %s (%s): %w", g.Name, g.Command, err)
	}

	values := make(Values)
	for _, line := range strings.Split(stdout, "\n") {
		// key, value, timestamp
		xs := strings.Fields(line)
		if len(xs) < 3 {
			continue
		}
		value, err := strconv.ParseFloat(xs[1], 64)
		if err != nil {
			logger.Warningf("plugin %s (%s): failed to parse value: %s", g.Name, g.Command, err)
			continue
		}
		values[pluginPrefix+xs[0]] = value
	}

	return values, nil
}

type pluginMeta struct {
	Graphs map[string]struct {
		Label   string
		Unit    string
		Metrics []struct {
			Name    string
			Label   string
			Stacked bool
		}
	}
}

// GetGraphDefs gets graph definitions
func (g *pluginGenerator) GetGraphDefs(ctx context.Context) ([]*mackerel.GraphDefsParam, error) {
	env := append(g.Env, pluginMetaEnvName+"=1")
	stdout, stderr, _, err := cmdutil.RunCommand(ctx, g.Command, g.User, env, g.Timeout)

	if stderr != "" {
		logger.Infof("plugin %s (%s): %q", g.Name, g.Command, stderr)
	}
	if err != nil {
		return nil, fmt.Errorf("plugin %s (%s): %w", g.Name, g.Command, err)
	}

	xs := strings.SplitN(stdout, "\n", 2)
	if len(xs) < 2 || !strings.HasPrefix(xs[0], pluginMetaHeadline) {
		logger.Infof("plugin %s: invalid plugin meta output: %q", g.Name, stdout)
		return nil, nil
	}

	var conf pluginMeta
	if err = json.Unmarshal([]byte(xs[1]), &conf); err != nil {
		return nil, fmt.Errorf("plugin %s: failed to decode plugin meta: %w", g.Name, err)
	}

	var graphDefs []*mackerel.GraphDefsParam
	for key, graph := range conf.Graphs {
		graphDef := mackerel.GraphDefsParam{
			Name:        pluginPrefix + key,
			DisplayName: graph.Label,
			Unit:        graph.Unit,
		}
		if graphDef.Unit == "" {
			graphDef.Unit = "float"
		}
		for _, metric := range graph.Metrics {
			graphDef.Metrics = append(graphDef.Metrics, &mackerel.GraphDefsMetric{
				Name:        pluginPrefix + key + "." + metric.Name,
				DisplayName: metric.Label,
				IsStacked:   metric.Stacked,
			})
		}
		graphDefs = append(graphDefs, &graphDef)
	}

	return graphDefs, nil
}
