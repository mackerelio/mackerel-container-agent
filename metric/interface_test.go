package metric

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestInterfaceGenerator(t *testing.T) {
	ctx := context.Background()
	generator := NewInterfaceGenerator()
	values, err := generator.Generate(ctx)
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	if values != nil {
		t.Errorf("should not generate values")
	}

	time.Sleep(time.Second * 1)

	values, err = generator.Generate(ctx)
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	if values == nil || len(values) == 0 {
		t.Errorf("should generate values")
	}

	for name := range values {
		if strings.HasPrefix(name, "interface.veth") {
			t.Errorf("value for %s should not generate values", name)
		}
	}
}
