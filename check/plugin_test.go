package check

import (
	"context"
	"strconv"
	"testing"
	"time"

	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/cmdutil"
	"github.com/mackerelio/mackerel-container-agent/config"
)

func TestPlugin_Generate(t *testing.T) {
	ctx := context.Background()
	g := NewPluginGenerator(&config.CheckPlugin{
		Name:    "dice",
		Command: cmdutil.CommandString("../example/check-dice.sh"),
	})
	result, err := g.Generate(ctx)
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	if expected := "dice"; result.name != expected {
		t.Errorf("name should be %q but got: %q", expected, result.name)
	}
	n, err := strconv.Atoi(result.message)
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	if n == 6 {
		if expected := mackerel.CheckStatusCritical; result.status != expected {
			t.Errorf("status should be %v but got: %v", expected, result.status)
		}
	} else if n == 4 || n == 5 {
		if expected := mackerel.CheckStatusWarning; result.status != expected {
			t.Errorf("status should be %v but got: %v", expected, result.status)
		}
	} else if 1 <= n && n <= 3 {
		if expected := mackerel.CheckStatusOK; result.status != expected {
			t.Errorf("status should be %v but got: %v", expected, result.status)
		}
	} else {
		t.Errorf("unexpected message: %v", result.message)
	}
}

func TestPlugin_Generate_Unknown(t *testing.T) {
	ctx := context.Background()
	g := NewPluginGenerator(&config.CheckPlugin{
		Name:    "unknown",
		Command: cmdutil.CommandString("../example/check-unknown.sh"),
	})
	result, err := g.Generate(ctx)
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	if expected := "unknown"; result.name != expected {
		t.Errorf("name should be %q but got: %q", expected, result.name)
	}
	if expected := mackerel.CheckStatusUnknown; result.status != expected {
		t.Errorf("status should be %v but got: %v", expected, result.status)
	}
}

func TestPlugin_Generate_Timeout(t *testing.T) {
	ctx := context.Background()
	g := NewPluginGenerator(&config.CheckPlugin{
		Name:    "timeout",
		Command: cmdutil.CommandString("sleep 10"),
		Timeout: 10 * time.Millisecond,
	})
	result, err := g.Generate(ctx)
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	if expected := "timeout"; result.name != expected {
		t.Errorf("name should be %q but got: %q", expected, result.name)
	}
	if expected := mackerel.CheckStatusUnknown; result.status != expected {
		t.Errorf("status should be %v but got: %v", expected, result.status)
	}
	if expected := "command timed out"; result.message != expected {
		t.Errorf("message should be %v but got: %v", expected, result.message)
	}
}

func TestPlugin_Generate_ok_ok(t *testing.T) {
	ctx := context.Background()
	g := NewPluginGenerator(&config.CheckPlugin{
		Name:    "ok",
		Command: cmdutil.CommandString("printf ok"),
	})
	result, err := g.Generate(ctx)
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	if expected := "ok"; result.name != expected {
		t.Errorf("name should be %q but got: %q", expected, result.name)
	}
	if expected := mackerel.CheckStatusOK; result.status != expected {
		t.Errorf("status should be %v but got: %v", expected, result.status)
	}

	result, err = g.Generate(ctx)
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	if result != nil {
		t.Errorf("result should be nil but got: %v", result)
	}
}

func TestPlugin_Generate_Env(t *testing.T) {
	ctx := context.Background()
	g := NewPluginGenerator(&config.CheckPlugin{
		Name:    "ok",
		Command: cmdutil.CommandString("printf '%s %s' $ENV2 $ENV1"),
		Env:     []string{"ENV1=foo", "ENV2=bar"},
	})
	result, err := g.Generate(ctx)
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	if expected := "ok"; result.name != expected {
		t.Errorf("name should be %q but got: %q", expected, result.name)
	}
	if expected := mackerel.CheckStatusOK; result.status != expected {
		t.Errorf("status should be %v but got: %v", expected, result.status)
	}
	if expected := "bar foo"; result.message != expected {
		t.Errorf("message should be %v but got: %v", expected, result.message)
	}
}
