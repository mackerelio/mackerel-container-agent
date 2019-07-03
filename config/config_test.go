package config

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"sort"
	"testing"
	"time"

	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/cmdutil"
)

var sampleConfig = `
apikey: 'DUMMY APIKEY'
`

func TestLoadDefault(t *testing.T) {
	os.Clearenv()
	conf, err := load("")
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	if !reflect.DeepEqual(conf, defaultConfig()) {
		t.Errorf("expect %#v, got %#v", defaultConfig(), conf)
	}
}

func TestLoadFile(t *testing.T) {
	file, err := newConfigFile(`
apikey: 'DUMMY APIKEY'
root: '/tmp/mackerel-container-agent'
`)
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	defer os.Remove(file.Name())

	expect := &Config{
		Apibase: "",
		Apikey:  "DUMMY APIKEY",
		Root:    "/tmp/mackerel-container-agent",
	}

	conf, err := load(file.Name())
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	if !reflect.DeepEqual(conf, expect) {
		t.Errorf("expect %#v, got %#v", expect, conf)
	}
}

func TestLoadHTTP(t *testing.T) {
	ts := newHTTPServer(sampleConfig)
	defer ts.Close()

	expect := &Config{
		Apibase: "",
		Apikey:  "DUMMY APIKEY",
		Root:    "/var/tmp/mackerel-container-agent",
	}

	conf, err := load(ts.URL)
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	if !reflect.DeepEqual(conf, expect) {
		t.Errorf("expect %#v, got %#v", expect, conf)
	}
}

func TestLoadS3(t *testing.T) {
	orgS3downloader := s3downloader
	s3downloader = newS3Downloader(sampleConfig)
	defer func() {
		s3downloader = orgS3downloader
	}()

	expect := &Config{
		Apibase: "",
		Apikey:  "DUMMY APIKEY",
		Root:    "/var/tmp/mackerel-container-agent",
	}

	conf, err := load("s3://bucket/key")
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	if !reflect.DeepEqual(conf, expect) {
		t.Errorf("expect %#v, got %#v", expect, conf)
	}
}

func TestLoadWithEnv(t *testing.T) {
	conf, err := newConfigFile(`
apibase: http://localhost:8080
apikey: 'DUMMY APIKEY'
`)
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	defer os.Remove(conf.Name())

	os.Setenv("MACKEREL_APIKEY", "ENV APIKEY")
	defer os.Unsetenv("MACKEREL_APIKEY")
	os.Setenv("MACKEREL_APIBASE", "http://127.0.0.1:9000")
	defer os.Unsetenv("MACKEREL_APIBASE")

	testCases := []struct {
		name     string
		location string
		expect   *Config
	}{
		{
			name:     "env",
			location: "",
			expect: &Config{
				Apibase: "http://127.0.0.1:9000",
				Apikey:  "ENV APIKEY",
				Root:    "/var/tmp/mackerel-container-agent",
			},
		},
		{
			name:     "config",
			location: conf.Name(),
			expect: &Config{
				Apibase: "http://localhost:8080",
				Apikey:  "DUMMY APIKEY",
				Root:    "/var/tmp/mackerel-container-agent",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conf, err := load(tc.location)
			if err != nil {
				t.Errorf("should not raise error: %v", err)
			}
			if !reflect.DeepEqual(conf, tc.expect) {
				t.Errorf("expect %#v, got %#v", tc.expect, conf)
			}
		})
	}
}

func TestRoles(t *testing.T) {
	conf, err := newConfigFile(`
roles:
  - My-Service:app
  - Another-Service:db
`)
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	defer os.Remove(conf.Name())

	os.Setenv("MACKEREL_ROLES", "Foo:xxx, Bar:yyy, Buz:zzz")
	defer os.Unsetenv("MACKEREL_ROLES")

	testCases := []struct {
		name     string
		location string
		expect   *Config
	}{
		{
			name:     "config",
			location: conf.Name(),
			expect: &Config{
				Root:  defaultRoot,
				Roles: []string{"My-Service:app", "Another-Service:db"},
			},
		},
		{
			name:     "env",
			location: "",
			expect: &Config{
				Root:  defaultRoot,
				Roles: []string{"Foo:xxx", "Bar:yyy", "Buz:zzz"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conf, err := load(tc.location)
			if err != nil {
				t.Errorf("should not raise error: %v", err)
			}
			if !reflect.DeepEqual(conf, tc.expect) {
				t.Errorf("expect %#v, got %#v", tc.expect, conf)
			}
		})
	}
}

func TestIgnoreContainer(t *testing.T) {
	conf, err := newConfigFile(`
ignoreContainer: "^mackerel-.+$"
`)
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	defer os.Remove(conf.Name())

	os.Setenv("MACKEREL_IGNORE_CONTAINER", "^mackerel-.+$")
	defer os.Unsetenv("MACKEREL_IGNORE_CONTAINER")

	r, _ := regexp.Compile("^mackerel-.+$")

	testCases := []struct {
		name     string
		location string
		expect   *Config
	}{
		{
			name:     "config",
			location: conf.Name(),
			expect: &Config{
				Root:            defaultRoot,
				IgnoreContainer: Regexpwrapper{r},
			},
		},
		{
			name:     "env",
			location: "",
			expect: &Config{
				Root:            defaultRoot,
				IgnoreContainer: Regexpwrapper{r},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conf, err := load(tc.location)
			if err != nil {
				t.Errorf("should not raise error: %v", err)
			}
			if !reflect.DeepEqual(conf, tc.expect) {
				t.Errorf("expect %#v, got %#v", tc.expect, conf)
			}
		})
	}
}

func TestMetricPlugins(t *testing.T) {
	file, err := newConfigFile(`
plugin:
  metrics:
    mysql:
      command: mackerel-plugin-mysql

    redis6379:
      command: mackerel-plugin-redis -port=6379 -timeout=5 -metric-key-prefix=redis6379
      timeoutSeconds: 50

    sample:
      command: ruby /usr/local/bin/sample-plugin.rb
      user: "sample-user"
      env:
        FOO: "FOO BAR"
        QUX: 'QUX QUUX'

    sample-args:
      command:
        - ruby
        - -arg0
        - 30
        - /usr/local/bin/sample-plugin.rb

  checks:
    procs:
      command: "check-procs --pattern=/usr/sbin/sshd --warning-under=1"
      user: "sample-user"
      env:
        FOO: FOO BAR
      timeoutSeconds: 45
      memo: "check procs memo"
`)
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	defer os.Remove(file.Name())

	expect := &Config{
		Apibase: "",
		Apikey:  "",
		Root:    defaultRoot,
		MetricPlugins: []*MetricPlugin{
			&MetricPlugin{
				Name:    "mysql",
				Command: cmdutil.CommandString("mackerel-plugin-mysql"),
			},
			&MetricPlugin{
				Name:    "redis6379",
				Command: cmdutil.CommandString("mackerel-plugin-redis -port=6379 -timeout=5 -metric-key-prefix=redis6379"),
				Timeout: 50 * time.Second,
			},
			&MetricPlugin{
				Name:    "sample",
				Command: cmdutil.CommandString("ruby /usr/local/bin/sample-plugin.rb"),
				User:    "sample-user",
				Env:     []string{"FOO=FOO BAR", "QUX=QUX QUUX"},
			},
			&MetricPlugin{
				Name:    "sample-args",
				Command: cmdutil.CommandArgs([]string{"ruby", "-arg0", "30", "/usr/local/bin/sample-plugin.rb"}),
			},
		},
		CheckPlugins: []*CheckPlugin{
			&CheckPlugin{
				Name:    "procs",
				Command: cmdutil.CommandString("check-procs --pattern=/usr/sbin/sshd --warning-under=1"),
				User:    "sample-user",
				Env:     []string{"FOO=FOO BAR"},
				Timeout: 45 * time.Second,
				Memo:    "check procs memo",
			},
		},
	}

	conf, err := load(file.Name())
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	sort.Slice(conf.MetricPlugins, func(i, j int) bool {
		return conf.MetricPlugins[i].Name < conf.MetricPlugins[j].Name
	})
	sort.Slice(conf.MetricPlugins[2].Env, func(i, j int) bool {
		return conf.MetricPlugins[2].Env[i] < conf.MetricPlugins[2].Env[j]
	})
	if !reflect.DeepEqual(conf, expect) {
		t.Errorf("expect %#v, got %#v", expect, conf)
	}
}

func TestReadinessProbe(t *testing.T) {
	proxy, _ := url.Parse("http://proxy.example.com:8080")

	testCases := []struct {
		name      string
		config    string
		expect    *Config
		shouldErr bool
	}{
		{
			name: "exec probe",
			config: `
readinessProbe:
  exec:
    command: cat /tmp/healthy
`,
			expect: &Config{
				Root: defaultRoot,
				ReadinessProbe: &Probe{
					Exec: &ProbeExec{
						Command: cmdutil.CommandString("cat /tmp/healthy"),
					},
				},
			},
		},
		{
			name: "exec probe command array",
			config: `
readinessProbe:
  exec:
    command:
      - cat
      - /tmp/healthy
`,
			expect: &Config{
				Root: defaultRoot,
				ReadinessProbe: &Probe{
					Exec: &ProbeExec{
						Command: cmdutil.CommandArgs([]string{"cat", "/tmp/healthy"}),
					},
				},
			},
		},
		{
			name: "exec probe with user and env",
			config: `
readinessProbe:
  exec:
    command: cat /tmp/healthy
    user: "sample-user"
    env:
      FOO: "FOO BAR"
`,
			expect: &Config{
				Root: defaultRoot,
				ReadinessProbe: &Probe{
					Exec: &ProbeExec{
						Command: cmdutil.CommandString("cat /tmp/healthy"),
						User:    "sample-user",
						Env:     []string{"FOO=FOO BAR"},
					},
				},
			},
		},
		{
			name: "exec probe error",
			config: `
readinessProbe:
  exec: {}
`,
			shouldErr: true,
		},
		{
			name: "http probe",
			config: `
readinessProbe:
  http:
    path: /healthy
`,
			expect: &Config{
				Root: defaultRoot,
				ReadinessProbe: &Probe{
					HTTP: &ProbeHTTP{
						Path: "/healthy",
					},
				},
			},
		},
		{
			name: "http probe scheme, method, etc.",
			config: `
readinessProbe:
  http:
    scheme: http
    method: PUT
    host: example.com
    port: 8080
    path: /healthy
    headers:
      - name: X-Custom-Header
        value: test
      - name: Host
        value: example.com
    proxy: "http://proxy.example.com:8080"
  initialDelaySeconds: 10
  timeoutSeconds: 5
  periodSeconds: 3
`,
			expect: &Config{
				Root: defaultRoot,
				ReadinessProbe: &Probe{
					HTTP: &ProbeHTTP{
						Scheme:  "http",
						Method:  "PUT",
						Host:    "example.com",
						Port:    "8080",
						Path:    "/healthy",
						Headers: []Header{{"X-Custom-Header", "test"}, {"Host", "example.com"}},
						Proxy:   URLWrapper{proxy},
					},
					InitialDelaySeconds: 10,
					TimeoutSeconds:      5,
					PeriodSeconds:       3,
				},
			},
		},
		{
			name: "http probe error",
			config: `
readinessProbe:
  http: {}
`,
			shouldErr: true,
		},
		{
			name: "tcp probe",
			config: `
readinessProbe:
  tcp:
    host: example.com
    port: 8080
`,
			expect: &Config{
				Root: defaultRoot,
				ReadinessProbe: &Probe{
					TCP: &ProbeTCP{
						Host: "example.com",
						Port: "8080",
					},
				},
			},
		},
		{
			name: "tcp probe error",
			config: `
readinessProbe:
  tcp: {}
`,
			shouldErr: true,
		},
		{
			name: "no probe error",
			config: `
readinessProbe:
  periodSeconds: 3
`,
			shouldErr: true,
		},
		{
			name: "multiple probes error (exec and http)",
			config: `
readinessProbe:
  exec:
    command: cat /tmp/healthy
  http:
    path: /healthy
  periodSeconds: 3
`,
			shouldErr: true,
		},
		{
			name: "multiple probes error (http and tcp)",
			config: `
readinessProbe:
  http:
    path: /healthy
  tcp:
    port: 8080
`,
			shouldErr: true,
		},
		{
			name: "invalid initialDelaySeconds",
			config: `
readinessProbe:
  exec:
    command: cat /tmp/healthy
  initialDelaySeconds: -1
`,
			shouldErr: true,
		},
		{
			name: "invalid periodSeconds",
			config: `
readinessProbe:
  exec:
    command: cat /tmp/healthy
  periodSeconds: -1
`,
			shouldErr: true,
		},
		{
			name: "invalid timeoutSeconds",
			config: `
readinessProbe:
  exec:
    command: cat /tmp/healthy
  timeoutSeconds: -1
`,
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			file, err := newConfigFile(tc.config)
			if err != nil {
				t.Fatalf("should not raise error: %v", err)
			}
			defer os.Remove(file.Name())

			conf, err := load(file.Name())
			if err != nil && !tc.shouldErr {
				t.Fatalf("should not raise error: %v", err)
			}
			if err == nil && tc.shouldErr {
				t.Fatalf("should raise error: %v", err)
			}
			if conf != nil && !reflect.DeepEqual(conf.ReadinessProbe, tc.expect.ReadinessProbe) {
				t.Errorf("expect %#v (exec: %#v, http:%#v, tcp:%#v), got %#v (exec: %#v, http:%#v, tcp:%#v)",
					tc.expect.ReadinessProbe, tc.expect.ReadinessProbe.Exec, tc.expect.ReadinessProbe.HTTP, tc.expect.ReadinessProbe.TCP,
					conf.ReadinessProbe, conf.ReadinessProbe.Exec, conf.ReadinessProbe.HTTP, conf.ReadinessProbe.TCP,
				)
			}
		})
	}
}

func TestHostStatusOnStart(t *testing.T) {
	testCases := []struct {
		name      string
		config    string
		env       string
		expect    *Config
		shouldErr bool
	}{
		{
			name: "working",
			config: `
hostStatusOnStart: working
`,
			expect: &Config{
				Root:              defaultRoot,
				HostStatusOnStart: mackerel.HostStatusWorking,
			},
		},
		{
			name: "standby",
			config: `
hostStatusOnStart: standby
`,
			expect: &Config{
				Root:              defaultRoot,
				HostStatusOnStart: mackerel.HostStatusStandby,
			},
		},
		{
			name: "maintenance",
			config: `
hostStatusOnStart: maintenance
`,
			expect: &Config{
				Root:              defaultRoot,
				HostStatusOnStart: mackerel.HostStatusMaintenance,
			},
		},
		{
			name: "poweroff",
			config: `
hostStatusOnStart: poweroff
`,
			expect: &Config{
				Root:              defaultRoot,
				HostStatusOnStart: mackerel.HostStatusPoweroff,
			},
		},
		{
			name: "error",
			config: `
hostStatusOnStart: unknown
`,
			shouldErr: true,
		},
		{
			name: "env",
			env:  "standby",
			expect: &Config{
				Root:              defaultRoot,
				HostStatusOnStart: mackerel.HostStatusStandby,
			},
		},
		{
			name: "config with env",
			config: `
hostStatusOnStart: working
`,
			env: "standby",
			expect: &Config{
				Root:              defaultRoot,
				HostStatusOnStart: mackerel.HostStatusWorking,
			},
		},
		{
			name:      "invalid env",
			env:       "unknown",
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("MACKEREL_HOST_STATUS_ON_START", tc.env)
			file, err := newConfigFile(tc.config)
			if err != nil {
				t.Fatalf("should not raise error: %v", err)
			}
			defer os.Remove(file.Name())

			conf, err := load(file.Name())
			if err != nil && !tc.shouldErr {
				t.Fatalf("should not raise error: %v", err)
			}
			if err == nil && tc.shouldErr {
				t.Fatalf("should raise error: %v", err)
			}
			if conf != nil && conf.HostStatusOnStart != tc.expect.HostStatusOnStart {
				t.Errorf("expect %#v, got %#v", tc.expect.HostStatusOnStart, conf.HostStatusOnStart)
			}
		})
	}
	os.Unsetenv("MACKEREL_HOST_STATUS_ON_START")
}

func newHTTPServer(content string) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(content))
	})
	return httptest.NewServer(handler)
}

func newConfigFile(content string) (*os.File, error) {
	temp, err := ioutil.TempFile("", "mackerel-config-test")
	if err != nil {
		return nil, err
	}
	if _, err := temp.WriteString(content); err != nil {
		os.Remove(temp.Name())
		return nil, err
	}
	temp.Sync()
	temp.Close()
	return temp, nil
}

type mockS3Downloader struct {
	content string
}

func (m *mockS3Downloader) download(u *url.URL) ([]byte, error) {
	return []byte(m.content), nil
}

func newS3Downloader(content string) downloader {
	return &mockS3Downloader{
		content: content,
	}
}
