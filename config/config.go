package config

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/cmdutil"
)

const (
	timeout     = 3 * time.Second
	defaultRoot = "/var/tmp/mackerel-container-agent"
)

// Config represents agent configuration
type Config struct {
	Apibase           string        `yaml:"apibase"`
	Apikey            string        `yaml:"apikey"`
	Root              string        `yaml:"root"`
	Roles             []string      `yaml:"roles"`
	DisplayName       string        `yaml:"displayName"`
	Memo              string        `yaml:"memo"`
	IgnoreContainer   Regexpwrapper `yaml:"ignoreContainer"`
	ReadinessProbe    *Probe        `yaml:"readinessProbe"`
	HostStatusOnStart HostStatus    `yaml:"hostStatusOnStart"`
	MetricPlugins     []*MetricPlugin
	CheckPlugins      []*CheckPlugin
}

// Regexpwrapper wraps regexp.Regexp
type Regexpwrapper struct {
	*regexp.Regexp
}

// UnmarshalText decodes regexp string
func (r *Regexpwrapper) UnmarshalText(text []byte) error {
	var err error
	r.Regexp, err = regexp.Compile(string(text))
	return err
}

// HostStatus represents host status
type HostStatus string

// UnmarshalText decodes host status string
func (s *HostStatus) UnmarshalText(text []byte) error {
	status := string(text)
	if status != mackerel.HostStatusWorking &&
		status != mackerel.HostStatusStandby &&
		status != mackerel.HostStatusMaintenance &&
		status != mackerel.HostStatusPoweroff {
		return fmt.Errorf("invalid host status: %q", status)
	}
	*s = HostStatus(status)
	return nil
}

func parseConfig(data []byte) (*Config, error) {
	var conf struct {
		Config `yaml:",inline"`
		Plugin map[string]map[string]struct {
			Command        cmdutil.Command `yaml:"command"`
			User           string          `yaml:"user"`
			TimeoutSeconds int             `yaml:"timeoutSeconds"`
			Env            Env             `yaml:"env"`
			Memo           string          `yaml:"memo"`
		} `yaml:"plugin"`
	}
	err := yaml.Unmarshal(data, &conf)
	if err != nil {
		return nil, err
	}
	for name, plugin := range conf.Plugin["metrics"] {
		if plugin.Command.IsEmpty() {
			return nil, errors.New("specify command of metric plugin")
		}
		conf.Config.MetricPlugins = append(conf.Config.MetricPlugins, &MetricPlugin{
			Name: name, Command: plugin.Command, User: plugin.User, Env: plugin.Env,
			Timeout: time.Duration(plugin.TimeoutSeconds) * time.Second,
		})
	}
	for name, plugin := range conf.Plugin["checks"] {
		if plugin.Command.IsEmpty() {
			return nil, errors.New("specify command of check plugin")
		}
		conf.Config.CheckPlugins = append(conf.Config.CheckPlugins, &CheckPlugin{
			Name: name, Command: plugin.Command, User: plugin.User, Env: plugin.Env,
			Timeout: time.Duration(plugin.TimeoutSeconds) * time.Second,
			Memo:    plugin.Memo,
		})
	}

	sort.Slice(conf.Config.MetricPlugins, func(i, j int) bool {
		return conf.Config.MetricPlugins[i].Name < conf.Config.MetricPlugins[j].Name
	})

	sort.Slice(conf.Config.CheckPlugins, func(i, j int) bool {
		return conf.Config.CheckPlugins[i].Name < conf.Config.CheckPlugins[j].Name
	})

	if conf.ReadinessProbe != nil {
		if err := conf.ReadinessProbe.validate(); err != nil {
			return nil, err
		}
	}
	return &conf.Config, nil
}

func load(ctx context.Context, location string) (*Config, error) {
	var conf *Config

	if location == "" {
		conf = defaultConfig()
	} else {
		data, err := fetch(ctx, location)
		if err != nil {
			return nil, err
		}

		conf, err = parseConfig(data)
		if err != nil {
			return nil, err
		}
	}

	if conf.Apibase == "" {
		conf.Apibase = os.Getenv("MACKEREL_APIBASE")
	}

	if conf.Apikey == "" {
		conf.Apikey = os.Getenv("MACKEREL_APIKEY")
	}

	if conf.Root == "" {
		conf.Root = defaultRoot
	}

	if v, ok := os.LookupEnv("MACKEREL_ROLES"); len(conf.Roles) == 0 && ok {
		conf.Roles = parseRoles(v)
	}

	if conf.DisplayName == "" {
		conf.DisplayName = os.Getenv("MACKEREL_DISPLAY_NAME")
	}

	if conf.Memo == "" {
		conf.Memo = os.Getenv("MACKEREL_MEMO")
	}

	if conf.IgnoreContainer.Regexp == nil {
		if r := os.Getenv("MACKEREL_IGNORE_CONTAINER"); r != "" {
			if err := conf.IgnoreContainer.UnmarshalText([]byte(r)); err != nil {
				return nil, err
			}
		}
	}

	if conf.HostStatusOnStart == "" {
		if s := os.Getenv("MACKEREL_HOST_STATUS_ON_START"); s != "" {
			if err := conf.HostStatusOnStart.UnmarshalText([]byte(s)); err != nil {
				return nil, err
			}
		}
	}

	return conf, nil
}

func fetch(ctx context.Context, location string) ([]byte, error) {
	u, err := url.Parse(location)
	if err != nil {
		return fetchFile(location)
	}

	switch u.Scheme {
	case "http", "https":
		return fetchHTTP(ctx, u)
	case "s3":
		return fetchS3(ctx, u)
	default:
		return fetchFile(u.Path)
	}
}

func fetchFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func fetchHTTP(ctx context.Context, u *url.URL) ([]byte, error) {
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	cl := http.Client{
		Timeout: timeout,
	}

	resp, err := cl.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func parseRoles(value string) []string {
	var roles []string
	for _, v := range strings.Split(value, ",") {
		roles = append(roles, strings.Trim(v, " "))
	}
	return roles
}
