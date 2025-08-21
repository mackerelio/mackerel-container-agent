package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/api"
)

type hostIDStore interface {
	load() (string, bool, error)
	remove() error
	save(id string) error
}

type hostResolver struct {
	client      api.Client
	hostIDStore hostIDStore
}

func newHostResolver(client api.Client, root string) *hostResolver {
	return &hostResolver{
		client:      client,
		hostIDStore: &hostIDFileStore{path: filepath.Join(root, "id")},
	}
}

func (r *hostResolver) getHost(hostParam *mackerel.CreateHostParam) (*mackerel.Host, bool, error) {
	var host *mackerel.Host
	hostID, notExist, err := r.getLocalHostID()
	if err != nil {
		if notExist {
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

func (r *hostResolver) getLocalHostID() (string, bool, error) {
	return r.hostIDStore.load()
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
	return r.hostIDStore.save(id)
}

func (r *hostResolver) removeHostID() error {
	return r.hostIDStore.remove()
}

type hostIDFileStore struct {
	path string
}

func (r *hostIDFileStore) load() (string, bool, error) {
	content, err := os.ReadFile(r.path)
	if err != nil {
		return "", os.IsNotExist(err), err
	}
	hostID := strings.TrimRight(string(content), "\r\n")
	if hostID == "" {
		return "", false, fmt.Errorf("host id file %s found but the content is empty", r.path)
	}
	return hostID, false, nil
}

func (r *hostIDFileStore) save(id string) error {
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

func (r *hostIDFileStore) remove() error {
	return os.Remove(r.path)
}

type hostIDMemoryStore struct {
	mu sync.Mutex

	id    string
	exist bool
}

func (r *hostIDMemoryStore) load() (string, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.exist {
		return r.id, false, nil
	} else {
		return "", true, fmt.Errorf("not initialized")
	}
}

func (r *hostIDMemoryStore) save(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.id = id
	r.exist = true

	return nil
}

func (r *hostIDMemoryStore) remove() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.id = ""
	r.exist = false

	return nil
}
