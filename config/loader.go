package config

// Loader represents a config loader
type Loader struct {
	location string
}

// NewLoader creates a new Loader
func NewLoader(location string) *Loader {
	return &Loader{location: location}
}

// Load loads agent configuration
func (l *Loader) Load() (*Config, error) {
	return Load(l.location)
}
