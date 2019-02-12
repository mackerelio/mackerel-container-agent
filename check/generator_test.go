package check

import (
	"errors"
	"time"

	mackerel "github.com/mackerelio/mackerel-client-go"
)

func createMockGenerators() []Generator {
	now := time.Now()
	g1 := NewMockGenerator("g1", "g1 memo", []*Result{
		NewResult("g1", "g1 ok", mackerel.CheckStatusOK, now),
	}, nil)
	g2 := NewMockGenerator("g2", "g2 memo", []*Result{
		NewResult("g2", "g2 ok", mackerel.CheckStatusOK, now),
		NewResult("g2", "g2 warning", mackerel.CheckStatusWarning, now.Add(time.Minute)),
		NewResult("g2", "g2 critical", mackerel.CheckStatusCritical, now.Add(2*time.Minute)),
		nil,
		NewResult("g2", "g2 ok", mackerel.CheckStatusOK, now.Add(4*time.Minute)),
	}, nil)
	g3 := NewMockGenerator(
		"g3", "g3 memo", nil, errors.New("failed to exec check plugin"),
	)
	return []Generator{g1, g2, g3}
}
