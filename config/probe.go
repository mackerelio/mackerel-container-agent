package config

import (
	"errors"

	"github.com/mackerelio/mackerel-container-agent/cmdutil"
)

// Probe configuration.
type Probe struct {
	Exec                *ProbeExec `yaml:"exec"`
	HTTP                *ProbeHTTP `yaml:"http"`
	TCP                 *ProbeTCP  `yaml:"tcp"`
	InitialDelaySeconds int        `yaml:"initialDelaySeconds"`
	PeriodSeconds       int        `yaml:"periodSeconds"`
	TimeoutSeconds      int        `yaml:"timeoutSeconds"`
}

func (p *Probe) validate() error {
	if p.Exec != nil && p.HTTP != nil || p.HTTP != nil && p.TCP != nil || p.TCP != nil && p.Exec != nil {
		return errors.New("either one of exec, http or tcp can be configured for probe")
	}
	if p.Exec == nil && p.HTTP == nil && p.TCP == nil {
		return errors.New("configure exec, http or tcp for probe")
	}
	if p.Exec != nil && p.Exec.Command.IsEmpty() {
		return errors.New("specify command of exec probe")
	}
	if p.HTTP != nil && p.HTTP.Path == "" {
		return errors.New("specify path of http probe")
	}
	if p.TCP != nil && p.TCP.Port == "" {
		return errors.New("specify port of tcp probe")
	}
	if p.InitialDelaySeconds < 0 {
		return errors.New("initialDelaySeconds should be positive")
	}
	if p.PeriodSeconds < 0 {
		return errors.New("periodSeconds should be positive")
	}
	if p.TimeoutSeconds < 0 {
		return errors.New("timeoutSeconds should be positive")
	}
	return nil
}

// ProbeExec is a probe with command.
type ProbeExec struct {
	Command cmdutil.Command `yaml:"command"`
	User    string          `yaml:"user"`
	Env     Env             `yaml:"env"`
}

// ProbeHTTP is a probe with http.
type ProbeHTTP struct {
	Scheme    string   `yaml:"scheme"`
	Method    string   `yaml:"method"`
	Host      string   `yaml:"host"`
	Port      string   `yaml:"port"`
	Path      string   `yaml:"path"`
	Headers   []Header `yaml:"headers"`
	UserAgent string
}

// Header is a request header for http probe.
type Header struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

// ProbeTCP is a probe with tcp.
type ProbeTCP struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}
