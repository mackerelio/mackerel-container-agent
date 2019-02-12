package check

import (
	"sync"

	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/api"
)

const maxPendingReports = 60

type sender struct {
	client         api.Client
	hostID         string
	pendingReports [][]*mackerel.CheckReport
	mu             sync.Mutex
}

func newSender(client api.Client) *sender {
	return &sender{client: client}
}

func (s *sender) post(reports []*mackerel.CheckReport) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pendingReports = append(s.pendingReports, reports)
	if s.hostID == "" {
		return nil
	}
	var postReports []*mackerel.CheckReport
	var postIndex int
	for i, r := range s.pendingReports {
		postIndex = i
		postReports = append(postReports, r...)
		if i > 1 {
			break
		}
	}
	for _, r := range postReports {
		r.Source = mackerel.NewCheckSourceHost(s.hostID)
	}
	var err error
	if len(postReports) > 0 {
		err = s.client.PostCheckReports(&mackerel.CheckReports{Reports: postReports})
	}
	if err == nil {
		n := copy(s.pendingReports, s.pendingReports[postIndex+1:])
		s.pendingReports = s.pendingReports[:n]
	} else {
		logger.Warningf("failed to post check monitoring reports but will retry posting: %s", err)
	}
	if len(s.pendingReports) > maxPendingReports {
		n := copy(s.pendingReports, s.pendingReports[len(s.pendingReports)-maxPendingReports:])
		s.pendingReports = s.pendingReports[:n]
	}
	return nil
}

func (s *sender) setHostID(hostID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hostID = hostID
}
