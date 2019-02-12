package metric

import (
	"sync"

	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/api"
)

type sender struct {
	client         api.Client
	hostID         string
	pendingMetrics [][]*mackerel.MetricValue
	mu             sync.Mutex
}

func newSender(client api.Client) *sender {
	return &sender{client: client}
}

func (s *sender) post(metricValues []*mackerel.MetricValue) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pendingMetrics = append(s.pendingMetrics, metricValues)
	if s.hostID == "" {
		return nil
	}
	var postMetricValues []*mackerel.MetricValue
	var postIndex int
	for i, ms := range s.pendingMetrics {
		postIndex = i
		postMetricValues = append(postMetricValues, ms...)
		if i > 1 { // send three oldest metrics at most
			break
		}
	}
	err := s.client.PostHostMetricValuesByHostID(s.hostID, postMetricValues)
	if err == nil {
		n := copy(s.pendingMetrics, s.pendingMetrics[postIndex+1:])
		s.pendingMetrics = s.pendingMetrics[:n]
	} else {
		logger.Warningf("failed to post metric values but will retry posting: %s", err)
	}
	if len(s.pendingMetrics) > 60*6 { // retry for 6 hours
		n := copy(s.pendingMetrics, s.pendingMetrics[len(s.pendingMetrics)-60*6:])
		s.pendingMetrics = s.pendingMetrics[:n]
	}
	return nil
}

func (s *sender) setHostID(hostID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hostID = hostID
}

func (s *sender) postGraphDefs(graphDefs []*mackerel.GraphDefsParam) error {
	if len(graphDefs) == 0 {
		return nil
	}
	return s.client.CreateGraphDefs(graphDefs)
}
