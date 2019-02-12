package spec

import (
	"sync"

	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/api"
)

type sender struct {
	client api.Client
	hostID string
	mu     sync.Mutex
}

func newSender(client api.Client) *sender {
	return &sender{client: client}
}

func (s *sender) post(param *mackerel.UpdateHostParam) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.hostID == "" { // skip updating host spec until host id is resolved
		return nil
	}
	_, err := s.client.UpdateHost(s.hostID, param)
	return err
}

func (s *sender) setHostID(hostID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hostID = hostID
}
