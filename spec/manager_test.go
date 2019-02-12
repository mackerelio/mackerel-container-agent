package spec

import (
	"context"
	"testing"
	"time"

	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/api"
)

func createMockSpecGenerators() []Generator {
	return []Generator{
		NewMockGenerator(nil, nil),
	}
}

func TestManagerRun(t *testing.T) {
	hostID := "abcde"
	var updatedCount int
	client := api.NewMockClient(
		api.MockUpdateHost(func(id string, param *mackerel.UpdateHostParam) (string, error) {
			updatedCount++
			if hostID != id {
				t.Fatal("inconsistent host id")
			}
			if expected := "mackerel-container-agent/x.y.z (Revision abc)"; param.Meta.AgentName != expected {
				t.Errorf("name should be %q but got: %q", expected, param.Meta.AgentName)
			}
			if expected := "x.y.z-container"; param.Meta.AgentVersion != expected {
				t.Errorf("version should be %q but got: %q", expected, param.Meta.AgentVersion)
			}
			if expected := "abc"; param.Meta.AgentRevision != expected {
				t.Errorf("revision should be %q but got: %q", expected, param.Meta.AgentRevision)
			}
			return hostID, nil
		}),
	)
	manager := NewManager(createMockSpecGenerators(), client).WithVersion("x.y.z", "abc")
	manager.SetHostID(hostID)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	err := manager.Run(ctx, 10*time.Millisecond, 100*time.Millisecond)
	if err != nil {
		t.Errorf("err should be nil but got: %+v", err)
	}
	if expected := 5; updatedCount != expected {
		t.Errorf("update host api is called %d times (expected: %d times)", updatedCount, expected)
	}
}

func TestManagerRun_LazyHostID(t *testing.T) {
	hostID := "abcde"
	var updatedCount int
	client := api.NewMockClient(
		api.MockUpdateHost(func(id string, param *mackerel.UpdateHostParam) (string, error) {
			updatedCount++
			if hostID != id {
				t.Fatal("inconsistent host id")
			}
			if expected := "mackerel-container-agent/x.y.z (Revision abc)"; param.Meta.AgentName != expected {
				t.Errorf("name should be %q but got: %q", expected, param.Meta.AgentName)
			}
			if expected := "x.y.z-container"; param.Meta.AgentVersion != expected {
				t.Errorf("version should be %q but got: %q", expected, param.Meta.AgentVersion)
			}
			if expected := "abc"; param.Meta.AgentRevision != expected {
				t.Errorf("revision should be %q but got: %q", expected, param.Meta.AgentRevision)
			}
			return hostID, nil
		}),
	)
	manager := NewManager(createMockSpecGenerators(), client).WithVersion("x.y.z", "abc")

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	go func() {
		time.Sleep(140 * time.Millisecond)
		manager.SetHostID(hostID)
	}()
	err := manager.Run(ctx, 10*time.Millisecond, 100*time.Millisecond)
	if err != nil {
		t.Errorf("err should be nil but got: %+v", err)
	}
	if expected := 3; updatedCount != expected {
		t.Errorf("update host api is called %d times (expected: %d times)", updatedCount, expected)
	}
}

func TestManagerRun_Hostname(t *testing.T) {
	hostID := "abcde"
	var updatedCount int
	client := api.NewMockClient(
		api.MockUpdateHost(func(id string, param *mackerel.UpdateHostParam) (string, error) {
			updatedCount++
			if hostID != id {
				t.Fatal("inconsistent host id")
			}
			if expected := "abcde012345"; param.Name != expected {
				t.Errorf("host name should be %q but got: %q", expected, param.Name)
			}
			return hostID, nil
		}),
	)
	manager := NewManager([]Generator{
		NewMockGenerator(&CloudHostname{
			Cloud:    nil,
			Hostname: "abcde012345",
		}, nil),
	}, client).WithVersion("x.y.z", "abc")
	manager.SetHostID(hostID)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	err := manager.Run(ctx, 10*time.Millisecond, 100*time.Millisecond)
	if err != nil {
		t.Errorf("err should be nil but got: %+v", err)
	}
}
