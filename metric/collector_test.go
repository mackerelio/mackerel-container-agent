package metric

import (
	"context"
	"reflect"
	"testing"
)

func TestCollectorCollect(t *testing.T) {
	ctx := context.Background()
	c := newCollector(createMockGenerators())
	values, err := c.collect(ctx)

	if err != nil {
		t.Errorf("error should be nil but got: %+v", err)
	}
	expectedValues := Values{
		"loadavg5":            2.39,
		"cpu.user.percentage": 29.2,
		"custom.foo.bar":      10.0,
		"custom.foo.baz":      20.0,
		"custom.foo.qux":      30.0,
		"custom.qux.a.bar":    12.39,
		"custom.qux.a.baz":    13.41,
		"custom.qux.b.bar":    14.43,
		"custom.qux.b.baz":    15.45,
	}
	if !reflect.DeepEqual(values, expectedValues) {
		t.Errorf("values should be %+v but got: %+v", expectedValues, values)
	}
}
