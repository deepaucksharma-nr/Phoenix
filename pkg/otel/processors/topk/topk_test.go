package topk

import (
	"testing"

	"github.com/phoenix/platform/pkg/otel"
)

func TestProcessor(t *testing.T) {
	factory := &Factory{}
	proc, err := factory.Create(map[string]interface{}{
		"metric_name": "cpu",
		"top_k":       2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	metrics := []otel.Metric{
		{Name: "cpu", Value: 3},
		{Name: "cpu", Value: 1},
		{Name: "cpu", Value: 2},
	}
	out := proc.Process(metrics)
	if len(out) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(out))
	}
	if out[0].Value != 3 || out[1].Value != 2 {
		t.Fatalf("unexpected ordering: %#v", out)
	}
}
