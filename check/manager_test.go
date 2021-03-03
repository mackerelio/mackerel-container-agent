package check

import (
	"context"
	"net/http"
	"reflect"
	"testing"
	"time"

	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/api"
)

func TestManager_Configs(t *testing.T) {
	client := api.NewMockClient()
	manager := NewManager(createMockGenerators(), client)
	expected := []mackerel.CheckConfig{
		{Name: "g1", Memo: "g1 memo"},
		{Name: "g2", Memo: "g2 memo"},
		{Name: "g3", Memo: "g3 memo"},
	}
	if !reflect.DeepEqual(expected, manager.Configs()) {
		t.Errorf("expected: %#v, got: %#v", expected, manager.Configs())
	}
}

func TestManagerRun(t *testing.T) {
	hostID := "abcde"
	var postedReports []*mackerel.CheckReports
	client := api.NewMockClient(
		api.MockCreateHost(func(param *mackerel.CreateHostParam) (string, error) {
			return hostID, nil
		}),
		api.MockPostCheckReports(func(reports *mackerel.CheckReports) error {
			postedReports = append(postedReports, reports)
			return nil
		}),
	)
	manager := NewManager(createMockGenerators(), client)
	manager.SetHostID(hostID)

	ctx, cancel := context.WithTimeout(context.Background(), 340*time.Millisecond)
	defer cancel()
	err := manager.Run(ctx, 50*time.Millisecond)
	if err != nil {
		t.Errorf("err should be nil but got: %+v", err)
	}
	if expected := 4; len(postedReports) != expected {
		t.Errorf("posted reports should have size %d but got: %d", expected, len(postedReports))
	}
	report := postedReports[2].Reports[0]
	if expected := "g2"; report.Name != expected {
		t.Errorf("report name should be %q but got: %q", expected, report.Name)
	}
	if expected := "g2 critical"; report.Message != expected {
		t.Errorf("report message should be %q but got: %q", expected, report.Message)
	}
	if expected := mackerel.CheckStatusCritical; report.Status != expected {
		t.Errorf("report status should be %v but got: %v", expected, report.Status)
	}
	if expected := mackerel.NewCheckSourceHost(hostID); !reflect.DeepEqual(report.Source, expected) {
		t.Errorf("report source should be %v but got: %v", expected, report.Source)
	}
}

func TestManagerRun_Retry(t *testing.T) {
	hostID := "abcde"
	var postedReports []*mackerel.CheckReports
	client := api.NewMockClient(
		api.MockCreateHost(func(param *mackerel.CreateHostParam) (string, error) {
			return hostID, nil
		}),
		api.MockPostCheckReports(func(reports *mackerel.CheckReports) error {
			return &mackerel.APIError{StatusCode: http.StatusInternalServerError}
		}),
	)
	go func() {
		time.Sleep(120 * time.Millisecond)
		client.ApplyOption(
			api.MockPostCheckReports(func(reports *mackerel.CheckReports) error {
				postedReports = append(postedReports, reports)
				return nil
			}),
		)
	}()
	manager := NewManager(createMockGenerators(), client)
	manager.SetHostID(hostID)

	// This test is flaky so we should sometimes retry.
	expectedNums := []int{4}
	for i := 0; i < 3; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 190*time.Millisecond)
		err := manager.Run(ctx, 50*time.Millisecond)
		cancel()
		if err != nil {
			t.Errorf("err should be nil but got: %+v", err)
			break
		}
		nums := reportCounts(postedReports)
		if reflect.DeepEqual(nums, expectedNums) {
			break
		}
		t.Logf("got %v; retry", nums)
	}
	nums := reportCounts(postedReports)
	if !reflect.DeepEqual(nums, expectedNums) {
		t.Errorf("posted reports should have size %v but got: %v", expectedNums, nums)
	}
}

func reportCounts(a []*mackerel.CheckReports) []int {
	nums := make([]int, len(a))
	for i, r := range a {
		nums[i] = len(r.Reports)
	}
	return nums
}
