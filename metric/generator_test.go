package metric

import (
	"errors"

	mackerel "github.com/mackerelio/mackerel-client-go"
)

func createMockGenerators() []Generator {
	g1 := NewMockGenerator(Values{
		"custom.foo.bar": 10.0,
		"custom.foo.baz": 20.0,
		"custom.foo.qux": 30.0,
	}, nil, []*mackerel.GraphDefsParam{
		&mackerel.GraphDefsParam{
			Name:        "custom.foo",
			DisplayName: "Foo graph",
			Unit:        "float",
			Metrics: []*mackerel.GraphDefsMetric{
				&mackerel.GraphDefsMetric{
					Name:        "custom.foo.bar",
					DisplayName: "Bar",
					IsStacked:   false,
				},
				&mackerel.GraphDefsMetric{
					Name:        "custom.foo.baz",
					DisplayName: "Baz",
					IsStacked:   false,
				},
				&mackerel.GraphDefsMetric{
					Name:        "custom.foo.qux",
					DisplayName: "Qux",
					IsStacked:   false,
				},
			},
		},
	}, nil)
	g2 := NewMockGenerator(Values{
		"custom.qux.a.bar": 12.39,
		"custom.qux.a.baz": 13.41,
		"custom.qux.b.bar": 14.43,
		"custom.qux.b.baz": 15.45,
	}, nil, []*mackerel.GraphDefsParam{
		&mackerel.GraphDefsParam{
			Name:        "custom.qux.#",
			DisplayName: "Qux graph",
			Unit:        "percentage",
			Metrics: []*mackerel.GraphDefsMetric{
				&mackerel.GraphDefsMetric{
					Name:        "custom.qux.#.bar",
					DisplayName: "Bar",
					IsStacked:   false,
				},
				&mackerel.GraphDefsMetric{
					Name:        "custom.qux.#.baz",
					DisplayName: "Baz",
					IsStacked:   false,
				},
			},
		},
	}, nil)
	g3 := NewMockGenerator(Values{
		"loadavg5":            2.39,
		"cpu.user.percentage": 29.2,
	}, nil, nil, nil)
	g4 := NewMockGenerator(
		Values{},
		errors.New("failed to fetch metrics"),
		nil,
		errors.New("failed to create graph definition"),
	)
	return []Generator{g1, g2, g3, g4}
}
