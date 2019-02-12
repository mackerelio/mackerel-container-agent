package check

import (
	"context"
	"time"

	mackerel "github.com/mackerelio/mackerel-client-go"
)

// Result represents check plugin result
type Result struct {
	name       string
	message    string
	status     mackerel.CheckStatus
	occurredAt time.Time
}

// NewResult creates a new Result
func NewResult(name string, message string, status mackerel.CheckStatus, occurredAt time.Time) *Result {
	return &Result{name, message, status, occurredAt}
}

// Generator interface generate check plugin result
type Generator interface {
	Generate(context.Context) (*Result, error)
	Config() mackerel.CheckConfig
}
