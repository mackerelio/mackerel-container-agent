package check

import (
	"context"
	"sync"

	mackerel "github.com/mackerelio/mackerel-client-go"
)

type collector struct {
	generators []Generator
}

func newCollector(generators []Generator) *collector {
	return &collector{
		generators: generators,
	}
}

func (c *collector) configs() []mackerel.CheckConfig {
	configs := make([]mackerel.CheckConfig, len(c.generators))
	for i, g := range c.generators {
		configs[i] = g.Config()
	}
	return configs
}

func (c *collector) collect(ctx context.Context) []*Result {
	var wg sync.WaitGroup
	reports := make([]*Result, 0, len(c.generators))
	mu := new(sync.Mutex)
	for _, g := range c.generators {
		wg.Add(1)
		go func(g Generator) {
			defer wg.Done()
			r, err := g.Generate(ctx)
			if err != nil {
				logger.Errorf("%s", err)
				return
			}
			mu.Lock()
			defer mu.Unlock()
			if r != nil {
				reports = append(reports, r)
			}
		}(g)
	}
	wg.Wait()
	return reports
}
