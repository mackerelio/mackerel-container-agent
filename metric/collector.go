package metric

import (
	"context"
	"maps"
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

func (c *collector) collect(ctx context.Context) (Values, error) {
	var wg sync.WaitGroup
	values := make(Values)
	mu := new(sync.Mutex)
	for _, g := range c.generators {
		wg.Go(func() {
			vs, err := g.Generate(ctx)
			if err != nil {
				logger.Errorf("%s", err)
				return
			}
			mu.Lock()
			defer mu.Unlock()
			maps.Copy(values, vs)
		})
	}
	wg.Wait()
	return values, nil
}

func (c *collector) collectGraphDefs(ctx context.Context) ([]*mackerel.GraphDefsParam, error) {
	var wg sync.WaitGroup
	var graphDefs []*mackerel.GraphDefsParam
	mu := new(sync.Mutex)
	for _, g := range c.generators {
		wg.Go(func() {
			gs, err := g.GetGraphDefs(ctx)
			if err != nil {
				logger.Errorf("%s", err)
				return
			}
			mu.Lock()
			defer mu.Unlock()
			graphDefs = append(graphDefs, gs...)
		})
	}
	wg.Wait()
	return graphDefs, nil
}
