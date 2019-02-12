package metric

import (
	"context"
	"testing"
	"time"

	"github.com/mackerelio/mackerel-container-agent/api"
)

func TestManagerRun(t *testing.T) {
	client := api.NewMockClient()
	hostID := "abcde"
	manager := NewManager(createMockGenerators(), client)

	ctx, cancel := context.WithTimeout(context.Background(), 190*time.Millisecond)
	defer cancel()
	go func() {
		time.Sleep(50 * time.Millisecond)
		if err := manager.CollectAndPostGraphDefs(ctx); err != nil {
			t.Errorf("err should be nil but got: %+v", err)
		}
		manager.SetHostID(hostID)
	}()
	err := manager.Run(ctx, 40*time.Millisecond)
	if err != nil {
		t.Errorf("err should be nil but got: %+v", err)
	}
	metricValues := client.PostedMetricValues()
	if expected := 9 * 4; len(metricValues[hostID]) != expected {
		t.Errorf("metric values should have size %d but got: %#v", expected, metricValues[hostID])
	}
	graphDefs := client.PostedGraphDefs()
	if expected := 2; len(graphDefs) != expected {
		t.Errorf("graph definitions should have size %d but got: %#v", expected, graphDefs)
	}
}
