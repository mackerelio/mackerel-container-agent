package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/api"
)

type hostResolver struct {
	path   string
	client api.Client
}

func newHostResolver(client api.Client, root string) *hostResolver {
	return &hostResolver{
		path:   filepath.Join(root, "id"),
		client: client,
	}
}

func (r *hostResolver) getHost(hostParam *mackerel.CreateHostParam) (*mackerel.Host, bool, error) {
	var host *mackerel.Host
	content, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			// host id file not found
			if hostParam.CustomIdentifier != "" {
				// find host from custom identifier
				hosts, err := r.client.FindHosts(&mackerel.FindHostsParam{
					CustomIdentifier: hostParam.CustomIdentifier,
					Statuses: []string{
						mackerel.HostStatusWorking,
						mackerel.HostStatusStandby,
						mackerel.HostStatusMaintenance,
						mackerel.HostStatusPoweroff,
					},
				})
				if err != nil {
					return nil, retryFromError(err), fmt.Errorf("failed to find host for custom identifier = %s: %w", hostParam.CustomIdentifier, err)
				}
				if len(hosts) > 0 {
					host = hosts[0]
					_, err = r.client.UpdateHost(host.ID, (*mackerel.UpdateHostParam)(hostParam))
					if err != nil {
						return nil, retryFromError(err), fmt.Errorf("failed to update host for id = %s: %w", host.ID, err)
					}
					return host, false, r.saveHostID(host.ID)
				}
			}
			// create a new host
			hostID, err := r.client.CreateHost(hostParam)
			if err != nil {
				return nil, retryFromError(err), fmt.Errorf("failed to create a new host: %w", err)
			}
			if err := r.saveHostID(hostID); err != nil {
				return nil, false, err
			}
			host, err = r.client.FindHost(hostID)
			if err != nil {
				return nil, retryFromError(err), fmt.Errorf("failed to find host for id = %s: %w", hostID, err)
			}
			return host, false, nil
		}
		return nil, false, err
	}
	hostID := strings.TrimRight(string(content), "\r\n")
	if hostID == "" {
		return nil, false, fmt.Errorf("host id file %s found but the content is empty", r.path)
	}
	host, err = r.client.FindHost(hostID)
	if err != nil {
		return nil, retryFromError(err), fmt.Errorf("failed to find host for id = %s: %w", hostID, err)
	}
	_, err = r.client.UpdateHost(host.ID, (*mackerel.UpdateHostParam)(hostParam))
	if err != nil {
		return nil, retryFromError(err), fmt.Errorf("failed to update host for id = %s: %w", hostID, err)
	}
	return host, false, nil
}

func (r *hostResolver) getLocalHostID() (string, error) {
	content, err := os.ReadFile(r.path)
	if err != nil {
		return "", err
	}
	hostID := strings.TrimRight(string(content), "\r\n")
	if hostID == "" {
		return "", fmt.Errorf("host id file found but the content is empty")
	}
	return hostID, nil
}

func retryFromError(err error) bool {
	if err == nil {
		return false
	}
	if err, ok := err.(*mackerel.APIError); ok {
		return err.StatusCode >= 500
	}
	return true
}

func (r *hostResolver) saveHostID(id string) error {
	err := os.MkdirAll(filepath.Dir(r.path), 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(r.path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write([]byte(id))
	if err != nil {
		return err
	}

	return nil
}

func (r *hostResolver) removeHostID() error {
	return os.Remove(r.path)
}
