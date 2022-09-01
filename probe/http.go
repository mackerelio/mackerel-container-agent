package probe

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mackerelio/mackerel-container-agent/config"
)

var (
	defaultTimeoutHTTP = 1 * time.Second
)

type probeHTTP struct {
	*config.ProbeHTTP
	initialDelay time.Duration
	period       time.Duration
	timeout      time.Duration
}

func (p *probeHTTP) Check(ctx context.Context) error {
	u, err := url.Parse(p.Path)
	if err != nil {
		return err
	}
	if u.Scheme = p.Scheme; u.Scheme == "" {
		u.Scheme = "http"
	}
	if p.Port != "" {
		host := p.Host
		if host == "" {
			host = "localhost"
		}
		u.Host = net.JoinHostPort(host, p.Port)
	} else if p.Host != "" {
		u.Host = p.Host
	} else {
		u.Host = "localhost"
	}

	client, err := p.createHTTPClient()
	if err != nil {
		return err
	}

	method := strings.ToUpper(p.Method)
	if method == "" {
		method = "GET"
	}
	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return err
	}
	for _, h := range p.Headers {
		if strings.ToLower(h.Name) == "host" {
			req.Host = h.Value
		} else {
			req.Header.Add(h.Name, h.Value)
		}
	}
	if req.Header.Get("User-Agent") == "" && p.UserAgent != "" {
		req.Header.Set("User-Agent", p.UserAgent)
	}

	res, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("http probe failed (%s %s): %w", method, u, err)
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || http.StatusBadRequest <= res.StatusCode {
		return fmt.Errorf("http probe failed (%s %s): %s", method, u, res.Status)
	}

	logger.Infof("http probe success (%s %s): %s", method, u, res.Status)
	return nil
}

func (p *probeHTTP) InitialDelay() time.Duration {
	return p.initialDelay
}

func (p *probeHTTP) Period() time.Duration {
	return p.period
}

func (p *probeHTTP) createHTTPClient() (*http.Client, error) {
	dt := http.DefaultTransport.(*http.Transport)
	tp := &http.Transport{
		DialContext:           dt.DialContext,
		MaxIdleConns:          dt.MaxIdleConns,
		IdleConnTimeout:       dt.IdleConnTimeout,
		TLSHandshakeTimeout:   dt.TLSHandshakeTimeout,
		ExpectContinueTimeout: dt.ExpectContinueTimeout,
		Proxy:                 http.ProxyURL(p.Proxy.URL),
	}

	timeout := p.timeout
	if timeout == 0 {
		timeout = defaultTimeoutHTTP
	}

	c := &http.Client{
		Timeout:   timeout,
		Transport: tp,
	}

	return c, nil
}
