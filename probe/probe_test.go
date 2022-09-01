package probe

import (
	"context"
	"errors"
	"testing"
	"time"
)

func init() {
	defaultPeriod = 100 * time.Millisecond
}

type mockProbe struct {
	results      []bool
	index        int
	count        int
	initialDelay time.Duration
	period       time.Duration
}

func newMockProbe(results []bool, initialDelay, period time.Duration) *mockProbe {
	return &mockProbe{
		results:      results,
		initialDelay: initialDelay,
		period:       period,
	}
}

func (p *mockProbe) Check(ctx context.Context) error {
	p.count++
	if p.index < len(p.results) {
		p.index++
	}
	if !p.results[p.index-1] {
		return errors.New("error")
	}
	return nil
}

func (p *mockProbe) InitialDelay() time.Duration {
	return p.initialDelay
}

func (p *mockProbe) Period() time.Duration {
	return p.period
}

func TestProbe_Wait(t *testing.T) {
	testCases := []struct {
		name         string
		results      []bool
		initialDelay time.Duration
		period       time.Duration
		count        int
		accuracy     int
		duration     time.Duration
	}{
		{
			name:     "ok",
			results:  []bool{true},
			count:    1,
			duration: time.Second,
		},
		{
			name:     "fail twice",
			results:  []bool{false, false, true},
			count:    3,
			duration: time.Second,
		},
		{
			name:     "stop by duration",
			results:  []bool{false},
			count:    3,
			accuracy: 1,
			duration: 250 * time.Millisecond,
		},
		{
			name:     "period",
			results:  []bool{false},
			period:   50 * time.Millisecond,
			count:    4,
			accuracy: 1,
			duration: 170 * time.Millisecond,
		},
		{
			name:         "initial delay",
			results:      []bool{false},
			initialDelay: 200 * time.Millisecond,
			count:        2,
			accuracy:     1,
			duration:     350 * time.Millisecond,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := newMockProbe(tc.results, tc.initialDelay, tc.period)
			ctx, cancel := context.WithTimeout(context.Background(), tc.duration)
			defer cancel()

			// Below writing a content sometimes expects returning an error.
			// The ctx may timeout while p waits to ready.
			Wait(ctx, p) // nolint

			// This test is flaky so we should check the count with an accuracy.
			if p.count < tc.count-tc.accuracy || p.count > tc.count+tc.accuracy {
				t.Errorf("Wait should check %d times with accuracy %d but got %d", tc.count, tc.accuracy, p.count)
			}
		})
	}
}
