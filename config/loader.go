package config

import (
	"context"
	"strconv"
	"time"
)

// Loader represents a config loader
type Loader struct {
	location        string
	pollingDuration time.Duration
}

// NewLoader creates a new Loader
func NewLoader(location, pollingDurationMinutes string) (*Loader, error) {
	var duration int
	var err error
	if pollingDurationMinutes != "" {
		duration, err = strconv.Atoi(pollingDurationMinutes)
		if err != nil {
			return nil, err
		}
	}
	return &Loader{
		location:        location,
		pollingDuration: time.Duration(duration) * time.Minute,
	}, nil
}

// Load loads agent configuration
func (l *Loader) Load() (*Config, error) {
	return load(l.location)
}

// Start the loader loop
func (l *Loader) Start(ctx context.Context) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		if l.pollingDuration > 0 {
			for {
				select {
				case <-time.After(l.pollingDuration):
					ch <- struct{}{}
					return
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return ch
}
