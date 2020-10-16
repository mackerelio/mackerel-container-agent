package hostinfo

import (
	"runtime"

	"github.com/mackerelio/go-osstat/memory"
)

// Generator interface gets host information
type Generator interface {
	Generate() (memTotal float64, cpuCores float64, err error)
}

// NewGenerator returns host info generator
func NewGenerator() Generator {
	return &generator{}
}

// generator is a real Generator
type generator struct{}

// Generate retrieves information and return it
func (r *generator) Generate() (memTotal float64, cpuCores float64, err error) {
	memory, err := memory.Get()
	if err != nil {
		return 0, 0, err
	}
	return float64(memory.Total), float64(runtime.NumCPU()), nil
}
