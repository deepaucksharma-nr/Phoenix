package adaptivefilter

import (
	"testing"

	"github.com/phoenix/platform/pkg/otel"
)

func TestProcessor(t *testing.T) {
	factory := &Factory{}
	proc, err := factory.Create(map[string]interface{}{
		"base_thresholds": map[string]interface{}{
			"cpu_percent": 2.0,
			"memory_mb":   100.0,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	metrics := []otel.Metric{
		{Name: "process.cpu.utilization", Value: 1},
		{Name: "process.cpu.utilization", Value: 5},
		{Name: "process.memory.usage", Value: 50},
		{Name: "process.memory.usage", Value: 200},
	}
	out := proc.Process(metrics)
	if len(out) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(out))
	}
}
