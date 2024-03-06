package agent

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/api"
	"github.com/mackerelio/mackerel-container-agent/check"
	"github.com/mackerelio/mackerel-container-agent/cmdutil"
	"github.com/mackerelio/mackerel-container-agent/config"
	"github.com/mackerelio/mackerel-container-agent/metric"
	"github.com/mackerelio/mackerel-container-agent/spec"
)

func createMockMetricGenerators() []metric.Generator {
	return []metric.Generator{
		metric.NewMockGenerator(metric.Values{
			"custom.foo.bar": 10.0,
			"custom.foo.baz": 20.0,
			"custom.foo.qux": 30.0,
		}, nil, []*mackerel.GraphDefsParam{
			&mackerel.GraphDefsParam{
				Name:        "custom.foo",
				DisplayName: "Foo graph",
				Unit:        "float",
				Metrics: []*mackerel.GraphDefsMetric{
					&mackerel.GraphDefsMetric{
						Name:        "custom.foo.bar",
						DisplayName: "Bar",
						IsStacked:   false,
					},
					&mackerel.GraphDefsMetric{
						Name:        "custom.foo.baz",
						DisplayName: "Baz",
						IsStacked:   false,
					},
					&mackerel.GraphDefsMetric{
						Name:        "custom.foo.qux",
						DisplayName: "Qux",
						IsStacked:   false,
					},
				},
			},
		}, nil),
	}
}

func createMockCheckGenerators() []check.Generator {
	return []check.Generator{
		check.NewMockGenerator("g1", "g1 memo", []*check.Result{
			check.NewResult("g1", "g1 ok", mackerel.CheckStatusOK, time.Now()),
		}, nil),
	}
}

func createMockSpecGenerators() []spec.Generator {
	return []spec.Generator{
		spec.NewMockGenerator(nil, nil),
	}
}

type mockPlatform struct{}

func (p *mockPlatform) GetMetricGenerators() []metric.Generator             { return nil }
func (p *mockPlatform) GetSpecGenerators() []spec.Generator                 { return nil }
func (p *mockPlatform) GetCustomIdentifier(context.Context) (string, error) { return "", nil }
func (p *mockPlatform) StatusRunning(context.Context) bool                  { return true }

func init() {
	metricsInterval = 200 * time.Millisecond
	checkInterval = 200 * time.Millisecond
	specInterval = 500 * time.Millisecond
	specInitialInterval = 600 * time.Millisecond
	waitStatusRunningInterval = 200 * time.Millisecond
	hostIDInitialRetryInterval = 100 * time.Millisecond
}

func TestAgentRun_RetryMetricPost(t *testing.T) {
	dir := t.TempDir()
	conf := &config.Config{Root: dir}
	hostID := "abcde"
	var postedMetricValues []*mackerel.MetricValue

	ctx, cancel := context.WithTimeout(context.Background(), 990*time.Millisecond)
	defer cancel()
	client := api.NewMockClient(
		api.MockCreateHost(func(param *mackerel.CreateHostParam) (string, error) {
			return hostID, nil
		}),
		api.MockFindHosts(func(param *mackerel.FindHostsParam) ([]*mackerel.Host, error) {
			return nil, errors.New("error")
		}),
		api.MockFindHost(func(id string) (*mackerel.Host, error) {
			return &mackerel.Host{ID: id}, nil
		}),
		api.MockPostHostMetricValuesByHostID(func(hostID string, metricValues []*mackerel.MetricValue) error {
			return &mackerel.APIError{StatusCode: http.StatusInternalServerError}
		}),
	)
	go func() {
		time.Sleep(500 * time.Millisecond)
		client.ApplyOption(
			api.MockPostHostMetricValuesByHostID(func(hostID string, metricValues []*mackerel.MetricValue) error {
				postedMetricValues = append(postedMetricValues, metricValues...)
				return nil
			}),
		)
	}()
	metricManager := metric.NewManager(createMockMetricGenerators(), client)
	checkManager := check.NewManager(createMockCheckGenerators(), client)
	specManager := spec.NewManager(createMockSpecGenerators(), client)

	_, err := run(ctx, client, metricManager, checkManager, specManager, &mockPlatform{}, conf)
	if err != nil {
		t.Errorf("err should be nil but got: %+v", err)
	}
	if expected := 3 * 4; len(postedMetricValues) != expected {
		t.Errorf("metric values should have size %d but got: %d", expected, len(postedMetricValues))
	}
}

func TestAgentRun_ResolveHostIdLazy(t *testing.T) {
	dir := t.TempDir()
	conf := &config.Config{Root: dir}
	hostID := "abcde"
	var updatedCount int
	var postedReports []*mackerel.CheckReports
	var createParam *mackerel.CreateHostParam
	var updateParam *mackerel.UpdateHostParam

	ctx, cancel := context.WithTimeout(context.Background(), 1900*time.Millisecond)
	defer cancel()
	client := api.NewMockClient(
		api.MockCreateHost(func(param *mackerel.CreateHostParam) (string, error) {
			return "", &mackerel.APIError{StatusCode: http.StatusInternalServerError}
		}),
		api.MockFindHosts(func(param *mackerel.FindHostsParam) ([]*mackerel.Host, error) {
			return nil, errors.New("error")
		}),
		api.MockFindHost(func(id string) (*mackerel.Host, error) {
			return &mackerel.Host{ID: id}, nil
		}),
		api.MockUpdateHost(func(id string, param *mackerel.UpdateHostParam) (string, error) {
			if id != hostID {
				return "", errors.New("invalid hostID")
			}
			updateParam = param
			updatedCount++
			return hostID, nil
		}),
		api.MockPostCheckReports(func(reports *mackerel.CheckReports) error {
			postedReports = append(postedReports, reports)
			return nil
		}),
	)
	go func() {
		time.Sleep(500 * time.Millisecond)
		client.ApplyOption(
			api.MockCreateHost(func(param *mackerel.CreateHostParam) (string, error) {
				createParam = param
				return hostID, nil
			}),
		)
	}()
	metricManager := metric.NewManager(createMockMetricGenerators(), client)
	checkManager := check.NewManager(createMockCheckGenerators(), client)
	specManager := spec.NewManager(createMockSpecGenerators(), client)

	_, err := run(ctx, client, metricManager, checkManager, specManager, &mockPlatform{}, conf)
	if err != nil {
		t.Errorf("err should be nil but got: %+v", err)
	}
	metricValues := client.PostedMetricValues()
	if expected := 3 * 9; len(metricValues[hostID]) != expected {
		t.Errorf("metric values should have size %d but got: %d", expected, len(metricValues[hostID]))
	}
	if expected := 2; updatedCount != expected {
		t.Errorf("update host api is called %d times (expected: %d times)", updatedCount, expected)
	}
	if expected := 1; len(postedReports) != expected {
		t.Errorf("posted reports should have size %d but got: %d", expected, len(postedReports))
	}
	if expected := "g1 ok"; postedReports[0].Reports[0].Message != expected {
		t.Errorf("posted report message should be %q but got: %q", expected, postedReports[0].Reports[0].Message)
	}
	if expected := []mackerel.CheckConfig{{Name: "g1", Memo: "g1 memo"}}; !reflect.DeepEqual(createParam.Checks, expected) {
		t.Errorf("expected: %#v got: %#v", expected, createParam.Checks)
	}
	if expected := []mackerel.CheckConfig{{Name: "g1", Memo: "g1 memo"}}; !reflect.DeepEqual(updateParam.Checks, expected) {
		t.Errorf("expected: %#v got: %#v", expected, updateParam.Checks)
	}
}

func TestAgentRun_NoRetryBadRequest(t *testing.T) {
	dir := t.TempDir()
	conf := &config.Config{Root: dir}
	hostID := "abcde"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client := api.NewMockClient(
		api.MockCreateHost(func(param *mackerel.CreateHostParam) (string, error) {
			return "", &mackerel.APIError{StatusCode: http.StatusBadRequest}
		}),
	)
	metricManager := metric.NewManager(createMockMetricGenerators(), client)
	checkManager := check.NewManager(createMockCheckGenerators(), client)
	specManager := spec.NewManager(createMockSpecGenerators(), client)

	_, err := run(ctx, client, metricManager, checkManager, specManager, &mockPlatform{}, conf)
	if err == nil {
		t.Errorf("err should not be nil but got: %+v", err)
	}
	metricValues := client.PostedMetricValues()
	if expected := 0; len(metricValues[hostID]) != expected {
		t.Errorf("metric values should have size %d but got: %d", expected, len(metricValues[hostID]))
	}
}

func TestAgentRun_Retire(t *testing.T) {
	dir := t.TempDir()
	conf := &config.Config{Root: dir}
	hostID := "abcde"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var retired bool
	client := api.NewMockClient(
		api.MockCreateHost(func(param *mackerel.CreateHostParam) (string, error) {
			return hostID, nil
		}),
		api.MockFindHost(func(id string) (*mackerel.Host, error) {
			return &mackerel.Host{ID: id}, nil
		}),
		api.MockRetireHost(func(id string) error {
			retired = true
			return nil
		}),
	)
	metricManager := metric.NewManager(createMockMetricGenerators(), client)
	checkManager := check.NewManager(createMockCheckGenerators(), client)
	specManager := spec.NewManager(createMockSpecGenerators(), client)

	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()
	retire, err := run(ctx, client, metricManager, checkManager, specManager, &mockPlatform{}, conf)
	if err != nil {
		t.Errorf("err should be nil but got: %+v", err)
	}
	retire()
	if !retired {
		t.Errorf("host should be retired")
	}
}

func TestAgentRun_Retire_Retry(t *testing.T) {
	dir := t.TempDir()
	conf := &config.Config{Root: dir}
	hostID := "abcde"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var retired bool
	client := api.NewMockClient(
		api.MockCreateHost(func(param *mackerel.CreateHostParam) (string, error) {
			return hostID, nil
		}),
		api.MockFindHost(func(id string) (*mackerel.Host, error) {
			return &mackerel.Host{ID: id}, nil
		}),
		api.MockRetireHost(func(id string) error {
			return &mackerel.APIError{StatusCode: http.StatusInternalServerError}
		}),
	)
	go func() {
		time.Sleep(1000 * time.Millisecond)
		client.ApplyOption(
			api.MockRetireHost(func(id string) error {
				retired = true
				return nil
			}),
		)
	}()
	metricManager := metric.NewManager(createMockMetricGenerators(), client)
	checkManager := check.NewManager(createMockCheckGenerators(), client)
	specManager := spec.NewManager(createMockSpecGenerators(), client)

	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()
	retire, err := run(ctx, client, metricManager, checkManager, specManager, &mockPlatform{}, conf)
	if err != nil {
		t.Errorf("err should be nil but got: %+v", err)
	}
	retire()
	if !retired {
		t.Errorf("host should be retired")
	}
}

func TestAgentRun_Retire_HostIDError(t *testing.T) {
	dir := t.TempDir()
	conf := &config.Config{Root: dir}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var retired bool
	client := api.NewMockClient(
		api.MockCreateHost(func(param *mackerel.CreateHostParam) (string, error) {
			return "", &mackerel.APIError{StatusCode: http.StatusInternalServerError}
		}),
		api.MockFindHost(func(id string) (*mackerel.Host, error) {
			return &mackerel.Host{ID: id}, nil
		}),
		api.MockRetireHost(func(id string) error {
			retired = true
			return nil
		}),
	)
	metricManager := metric.NewManager(createMockMetricGenerators(), client)
	checkManager := check.NewManager(createMockCheckGenerators(), client)
	specManager := spec.NewManager(createMockSpecGenerators(), client)

	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()
	retire, err := run(ctx, client, metricManager, checkManager, specManager, &mockPlatform{}, conf)
	if err != nil {
		t.Errorf("err should be nil but got: %+v", err)
	}
	retire() // fails because host id is not resolved
	if retired {
		t.Errorf("host should not be retired")
	}
}

func TestAgentRun_MetricPlugin(t *testing.T) {
	dir := t.TempDir()
	conf := &config.Config{
		Root: dir,
		MetricPlugins: []*config.MetricPlugin{
			&config.MetricPlugin{
				Name:    "dice",
				Command: cmdutil.CommandString("../example/dice.sh"),
			},
		},
	}
	hostID := "abcde"
	var postedMetricValues []*mackerel.MetricValue

	ctx, cancel := context.WithTimeout(context.Background(), 950*time.Millisecond)
	defer cancel()
	client := api.NewMockClient(
		api.MockCreateHost(func(param *mackerel.CreateHostParam) (string, error) {
			return hostID, nil
		}),
		api.MockFindHost(func(id string) (*mackerel.Host, error) {
			return &mackerel.Host{ID: id}, nil
		}),
		api.MockPostHostMetricValuesByHostID(func(hostID string, metricValues []*mackerel.MetricValue) error {
			postedMetricValues = append(postedMetricValues, metricValues...)
			return nil
		}),
	)
	var metricGenerators []metric.Generator
	for _, mp := range conf.MetricPlugins {
		metricGenerators = append(metricGenerators, metric.NewPluginGenerator(mp))
	}
	metricManager := metric.NewManager(metricGenerators, client)
	checkManager := check.NewManager(nil, client)
	specManager := spec.NewManager(createMockSpecGenerators(), client)

	_, err := run(ctx, client, metricManager, checkManager, specManager, &mockPlatform{}, conf)
	if err != nil {
		t.Errorf("err should be nil but got: %+v", err)
	}
	if expected := 2 * 4; len(postedMetricValues) != expected {
		t.Errorf("metric values should have size %d but got: %d", expected, len(postedMetricValues))
	}
	graphDefs := client.PostedGraphDefs()
	if expected := 1; len(graphDefs) != expected {
		t.Errorf("graph definitions should have size %d but got: %d", expected, len(graphDefs))
	}
	if expected := "custom.dice"; graphDefs[0].Name != expected {
		t.Errorf("expected: %#v, got: %#v", expected, graphDefs[0].Name)
	}
	if expected := "My Dice"; graphDefs[0].DisplayName != expected {
		t.Errorf("expected: %#v, got: %#v", expected, graphDefs[0].DisplayName)
	}
}

func TestAgentRun_CustomIdentifier(t *testing.T) {
	dir := t.TempDir()
	conf := &config.Config{Root: dir}
	hostID := "abcde"
	customIdentifier := "custom-identifier-abcde"
	var updatedCount int
	var postedMetricValues []*mackerel.MetricValue

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	client := api.NewMockClient(
		api.MockCreateHost(func(param *mackerel.CreateHostParam) (string, error) {
			return "", errors.New("error")
		}),
		api.MockFindHosts(func(param *mackerel.FindHostsParam) ([]*mackerel.Host, error) {
			return []*mackerel.Host{{ID: hostID}}, nil
		}),
		api.MockFindHost(func(id string) (*mackerel.Host, error) {
			return nil, errors.New("error")
		}),
		api.MockUpdateHost(func(id string, param *mackerel.UpdateHostParam) (string, error) {
			if id != hostID {
				return "", errors.New("invalid hostID")
			}
			updatedCount++
			return hostID, nil
		}),
		api.MockPostHostMetricValuesByHostID(func(id string, metricValues []*mackerel.MetricValue) error {
			if id != hostID {
				return errors.New("invalid hostID")
			}
			postedMetricValues = append(postedMetricValues, metricValues...)
			return nil
		}),
	)
	metricManager := metric.NewManager(createMockMetricGenerators(), client)
	checkManager := check.NewManager(createMockCheckGenerators(), client)
	specManager := spec.NewManager(createMockSpecGenerators(), client).WithCustomIdentifier(customIdentifier)

	_, err := run(ctx, client, metricManager, checkManager, specManager, &mockPlatform{}, conf)
	if err != nil {
		t.Errorf("err should be nil but got: %+v", err)
	}
	if expected := 1; updatedCount != expected {
		t.Errorf("update host api is called %d times (expected: %d times)", updatedCount, expected)
	}
	if expected := 3 * 2; len(postedMetricValues) != expected {
		t.Errorf("metric values should have size %d but got: %d", expected, len(postedMetricValues))
	}
}

func TestAgentRun_CustomIdentifier_CreateHost(t *testing.T) {
	dir := t.TempDir()
	conf := &config.Config{Root: dir}
	hostID := "abcde"
	customIdentifier := "custom-identifier-abcde"
	var postedMetricValues []*mackerel.MetricValue
	var updatedCount int

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	client := api.NewMockClient(
		api.MockCreateHost(func(param *mackerel.CreateHostParam) (string, error) {
			if param.CustomIdentifier != customIdentifier {
				return "", errors.New("invalid customIdentifier")
			}
			return hostID, nil
		}),
		api.MockFindHosts(func(param *mackerel.FindHostsParam) ([]*mackerel.Host, error) {
			return []*mackerel.Host{{ID: hostID}}, nil
		}),
		api.MockFindHost(func(id string) (*mackerel.Host, error) {
			return &mackerel.Host{ID: id}, nil
		}),
		api.MockUpdateHost(func(id string, param *mackerel.UpdateHostParam) (string, error) {
			if id != hostID {
				return "", errors.New("invalid hostID")
			}
			updatedCount++
			return hostID, nil
		}),
		api.MockPostHostMetricValuesByHostID(func(id string, metricValues []*mackerel.MetricValue) error {
			if id != hostID {
				return errors.New("invalid hostID")
			}
			postedMetricValues = append(postedMetricValues, metricValues...)
			return nil
		}),
	)
	metricManager := metric.NewManager(createMockMetricGenerators(), client)
	checkManager := check.NewManager(createMockCheckGenerators(), client)
	specManager := spec.NewManager(createMockSpecGenerators(), client).WithCustomIdentifier(customIdentifier)

	_, err := run(ctx, client, metricManager, checkManager, specManager, &mockPlatform{}, conf)
	if err != nil {
		t.Errorf("err should be nil but got: %+v", err)
	}
	if expected := 1; updatedCount != expected {
		t.Errorf("update host api is called %d times (expected: %d times)", updatedCount, expected)
	}
	if expected := 3 * 2; len(postedMetricValues) != expected {
		t.Errorf("metric values should have size %d but got: %d", expected, len(postedMetricValues))
	}
}

func TestAgentRun_HostIDFile(t *testing.T) {
	dir := t.TempDir()
	conf := &config.Config{Root: dir}
	hostID := "abcde"
	var updatedCount int
	var postedReports []*mackerel.CheckReports

	if err := os.WriteFile(filepath.Join(dir, "id"), []byte(hostID), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	client := api.NewMockClient(
		api.MockCreateHost(func(param *mackerel.CreateHostParam) (string, error) {
			return "", &mackerel.APIError{StatusCode: http.StatusInternalServerError}
		}),
		api.MockFindHosts(func(param *mackerel.FindHostsParam) ([]*mackerel.Host, error) {
			return nil, &mackerel.APIError{StatusCode: http.StatusInternalServerError}
		}),
		api.MockFindHost(func(id string) (*mackerel.Host, error) {
			return &mackerel.Host{ID: id}, nil
		}),
		api.MockUpdateHost(func(id string, param *mackerel.UpdateHostParam) (string, error) {
			if id != hostID {
				return "", errors.New("invalid hostID")
			}
			updatedCount++
			return hostID, nil
		}),
		api.MockPostCheckReports(func(reports *mackerel.CheckReports) error {
			postedReports = append(postedReports, reports)
			return nil
		}),
	)
	metricManager := metric.NewManager(createMockMetricGenerators(), client)
	checkManager := check.NewManager(createMockCheckGenerators(), client)
	specManager := spec.NewManager(createMockSpecGenerators(), client)

	_, err := run(ctx, client, metricManager, checkManager, specManager, &mockPlatform{}, conf)
	if err != nil {
		t.Errorf("err should be nil but got: %+v", err)
	}
	if expected := 1; updatedCount != expected {
		t.Errorf("update host api is called %d times (expected: %d times)", updatedCount, expected)
	}
}

type mockPlatformStatusRunning struct {
	count int
}

func (p *mockPlatformStatusRunning) GetMetricGenerators() []metric.Generator { return nil }
func (p *mockPlatformStatusRunning) GetSpecGenerators() []spec.Generator     { return nil }
func (p *mockPlatformStatusRunning) GetCustomIdentifier(context.Context) (string, error) {
	return "", nil
}
func (p *mockPlatformStatusRunning) StatusRunning(context.Context) bool {
	if p.count > 0 {
		p.count--
		return false
	}
	return true
}
func (p *mockPlatformStatusRunning) Status() string {
	if p.count > 0 {
		return "PENDNIG"
	}
	return "RUNNING"
}

type mockSpecGeneratorStatus struct {
	pform *mockPlatformStatusRunning
}

func (g *mockSpecGeneratorStatus) Generate(context.Context) (any, error) {
	return &mackerel.Cloud{
		MetaData: map[string]string{"status": g.pform.Status()},
	}, nil
}

func TestAgentRun_StatusRunning(t *testing.T) {
	dir := t.TempDir()
	conf := &config.Config{Root: dir}
	hostID := "abcde"
	pform := &mockPlatformStatusRunning{count: 2}

	ctx, cancel := context.WithTimeout(context.Background(), 700*time.Millisecond)
	defer cancel()
	client := api.NewMockClient(
		api.MockCreateHost(func(param *mackerel.CreateHostParam) (string, error) {
			expected := map[string]string{"status": "RUNNING"}
			if !reflect.DeepEqual(param.Meta.Cloud.MetaData, expected) {
				t.Errorf("expected: %#v got: %#v", expected, param.Meta.Cloud.MetaData)
			}
			return hostID, nil
		}),
		api.MockFindHost(func(id string) (*mackerel.Host, error) {
			return &mackerel.Host{ID: id}, nil
		}),
	)
	metricManager := metric.NewManager(createMockMetricGenerators(), client)
	checkManager := check.NewManager(createMockCheckGenerators(), client)
	specManager := spec.NewManager([]spec.Generator{&mockSpecGeneratorStatus{pform}}, client)

	_, err := run(ctx, client, metricManager, checkManager, specManager, pform, conf)
	if err != nil {
		t.Errorf("err should be nil but got: %+v", err)
	}
}

func TestAgentRun_ReadinessProbe(t *testing.T) {
	dir := t.TempDir()
	conf := &config.Config{
		Root: dir,
		ReadinessProbe: &config.Probe{
			Exec: &config.ProbeExec{
				Command: cmdutil.CommandString("sleep .3"),
			},
		},
	}
	hostID := "abcde"
	var postedMetricValues []*mackerel.MetricValue
	ctx, cancel := context.WithTimeout(context.Background(), 700*time.Millisecond)
	defer cancel()
	client := api.NewMockClient(
		api.MockCreateHost(func(param *mackerel.CreateHostParam) (string, error) {
			return hostID, nil
		}),
		api.MockFindHost(func(id string) (*mackerel.Host, error) {
			return &mackerel.Host{ID: id}, nil
		}),
		api.MockPostHostMetricValuesByHostID(func(hostID string, metricValues []*mackerel.MetricValue) error {
			postedMetricValues = append(postedMetricValues, metricValues...)
			return nil
		}),
	)
	metricManager := metric.NewManager(createMockMetricGenerators(), client)
	checkManager := check.NewManager(createMockCheckGenerators(), client)
	specManager := spec.NewManager(createMockSpecGenerators(), client)
	_, err := run(ctx, client, metricManager, checkManager, specManager, &mockPlatform{}, conf)
	if err != nil {
		t.Errorf("err should be nil but got: %+v", err)
	}
	if expected := 3 * 3; len(postedMetricValues) != expected {
		t.Errorf("metric values should have size %d but got: %d", expected, len(postedMetricValues))
	}
}

func TestAgentRun_ReadinessProbe_SleepLong(t *testing.T) {
	dir := t.TempDir()
	conf := &config.Config{
		Root: dir,
		ReadinessProbe: &config.Probe{
			Exec: &config.ProbeExec{
				Command: cmdutil.CommandString("sleep 5"),
			},
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	client := api.NewMockClient(
		api.MockCreateHost(func(param *mackerel.CreateHostParam) (string, error) {
			t.Errorf("create host should not be called")
			return "", nil
		}),
	)
	metricManager := metric.NewManager(createMockMetricGenerators(), client)
	checkManager := check.NewManager(createMockCheckGenerators(), client)
	specManager := spec.NewManager(createMockSpecGenerators(), client)
	_, err := run(ctx, client, metricManager, checkManager, specManager, &mockPlatform{}, conf)
	if err != nil {
		t.Errorf("err should be nil but got: %+v", err)
	}
}

func TestAgentRun_HostStatusOnStart(t *testing.T) {
	dir := t.TempDir()
	conf := &config.Config{
		Root:              dir,
		HostStatusOnStart: mackerel.HostStatusWorking,
	}
	hostID := "abcde"
	var postedStatus string

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	client := api.NewMockClient(
		api.MockCreateHost(func(param *mackerel.CreateHostParam) (string, error) {
			return hostID, nil
		}),
		api.MockFindHost(func(id string) (*mackerel.Host, error) {
			return &mackerel.Host{ID: id, Status: mackerel.HostStatusStandby}, nil
		}),
		api.MockUpdateHostStatus(func(id string, status string) error {
			if id != hostID {
				return errors.New("invalid hostID")
			}
			postedStatus = status
			return nil
		}),
	)
	metricManager := metric.NewManager(createMockMetricGenerators(), client)
	checkManager := check.NewManager(createMockCheckGenerators(), client)
	specManager := spec.NewManager(createMockSpecGenerators(), client)
	_, err := run(ctx, client, metricManager, checkManager, specManager, &mockPlatform{}, conf)
	if err != nil {
		t.Errorf("err should be nil but got: %+v", err)
	}
	if expected := "working"; postedStatus != expected {
		t.Errorf("posted host status should be %q but got: %q", expected, postedStatus)
	}
}
