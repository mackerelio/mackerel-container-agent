package metric

import (
	"context"
	"reflect"
	"testing"

	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/cmdutil"
	"github.com/mackerelio/mackerel-container-agent/config"
)

func TestPlugin_Generate(t *testing.T) {
	ctx := context.Background()
	g := NewPluginGenerator(&config.MetricPlugin{
		Name:    "dice",
		Command: cmdutil.CommandString("../example/dice.sh"),
	})
	values, err := g.Generate(ctx)
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	if len(values) != 2 {
		t.Errorf("values should have size 2 but got: %v", values)
	}
	value := values["custom.dice.d6"]
	if value < 1 || 6 < value {
		t.Errorf("dice.d6 should be 1 to 6 but got: %v", values)
	}
	value = values["custom.dice.d20"]
	if value < 1 || 20 < value {
		t.Errorf("dice.d20 should be 1 to 20 but got: %v", values)
	}
}

func TestPlugin_GetGraphDefs(t *testing.T) {
	ctx := context.Background()
	g := NewPluginGenerator(&config.MetricPlugin{
		Name:    "dice",
		Command: cmdutil.CommandString("../example/dice.sh"),
	})
	graphDefs, err := g.GetGraphDefs(ctx)
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	expectedGraphDefs := []*mackerel.GraphDefsParam{
		&mackerel.GraphDefsParam{
			Name:        "custom.dice",
			DisplayName: "My Dice",
			Unit:        "integer",
			Metrics: []*mackerel.GraphDefsMetric{
				&mackerel.GraphDefsMetric{
					Name:        "custom.dice.d6",
					DisplayName: "Die 6",
					IsStacked:   false,
				},
				&mackerel.GraphDefsMetric{
					Name:        "custom.dice.d20",
					DisplayName: "Die 20",
					IsStacked:   false,
				},
			},
		},
	}
	if !reflect.DeepEqual(graphDefs, expectedGraphDefs) {
		t.Errorf("expected: %#v, got: %#v", expectedGraphDefs, graphDefs)
	}
}

func TestPlugin_WithEnvGenerate(t *testing.T) {
	ctx := context.Background()
	g := NewPluginGenerator(&config.MetricPlugin{
		Name:    "dice",
		Command: cmdutil.CommandString("../example/env.sh"),
		Env:     []string{"NUM=128"},
	})
	values, err := g.Generate(ctx)
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	if len(values) != 1 {
		t.Errorf("values should have size 1 but got: %v", values)
	}
	value := values["custom.dice.d128"]
	if value < 1 || 128 < value {
		t.Errorf("dice.d128 should be 1 to 128 but got: %v", values)
	}
}

func TestPlugin_WithEnvGetGraphDefs(t *testing.T) {
	ctx := context.Background()
	g := NewPluginGenerator(&config.MetricPlugin{
		Name:    "dice",
		Command: cmdutil.CommandString("../example/env.sh"),
		Env:     []string{"NUM=128"},
	})
	graphDefs, err := g.GetGraphDefs(ctx)
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	if expected := "My Dice 128"; graphDefs[0].DisplayName != expected {
		t.Errorf("expected: %#v, got: %#v", expected, graphDefs[0].DisplayName)
	}
	if expected := "custom.dice.d128"; graphDefs[0].Metrics[0].Name != expected {
		t.Errorf("expected: %#v, got: %#v", expected, graphDefs[0].Metrics[0].Name)
	}
}
