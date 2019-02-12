package config

import (
	"time"

	"github.com/mackerelio/mackerel-container-agent/cmdutil"
)

// MetricPlugin represents metric plugin
type MetricPlugin struct {
	Name    string
	Command cmdutil.Command
	User    string
	Env     Env
	Timeout time.Duration
}

// CheckPlugin represents check plugin
type CheckPlugin struct {
	Name    string
	Command cmdutil.Command
	User    string
	Env     Env
	Timeout time.Duration
	Memo    string
}
