package spec

import (
	"bufio"
	"context"
	"io"
	"os"
	"strings"

	"github.com/mackerelio/golib/logging"
	mackerel "github.com/mackerelio/mackerel-client-go"
)

// CPUGenerator collects CPU specs
type CPUGenerator struct {
}

var cpuLogger = logging.GetLogger("spec.cpu")

func (g *CPUGenerator) generate(file io.Reader) (interface{}, error) {
	scanner := bufio.NewScanner(file)

	var results mackerel.CPU
	var cpuinfo map[string]interface{}
	var modelName string

	for scanner.Scan() {
		line := scanner.Text()
		kv := strings.SplitN(line, ":", 2)
		if len(kv) < 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])

		switch key {
		case "processor":
			cpuinfo = make(map[string]interface{})
			if modelName != "" {
				cpuinfo["model_name"] = modelName
			}
			results = append(results, cpuinfo)
		case "Processor", "system type":
			modelName = val
		case "vendor_id", "model", "stepping", "physical id", "core id", "model name", "cache size":
			cpuinfo[strings.Replace(key, " ", "_", -1)] = val
		case "cpu family":
			cpuinfo["family"] = val
		case "cpu cores":
			cpuinfo["cores"] = val
		case "cpu MHz":
			cpuinfo["mhz"] = val
		}
	}

	if err := scanner.Err(); err != nil {
		// Don't return error to prevent stop agent
		// caused by failing on scanning /proc/cpuinfo
		cpuLogger.Errorf("failed (on scanning /proc/cpuinfo): %s", err)
		return nil, nil
	}

	// Old kernels with CONFIG_SMP disabled has no "processor: " line
	if len(results) == 0 && modelName != "" {
		cpuinfo = make(map[string]interface{})
		cpuinfo["model_name"] = modelName
		results = append(results, cpuinfo)
	}

	return results, nil
}

// Generate CPU specs
func (g *CPUGenerator) Generate(ctx context.Context) (interface{}, error) {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		// Don't return error to prevent stop agent
		// caused by failing on opening /proc/cpuinfo
		cpuLogger.Errorf("failed (skip this spec): %s", err)
		return nil, nil
	}
	defer file.Close()

	return g.generate(file)
}
