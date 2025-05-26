package topk

import (
	"fmt"
	"sort"

	"github.com/phoenix/platform/pkg/otel"
)

// Config defines the processor configuration.
type Config struct {
	MetricName string
	TopK       int
}

// Processor keeps only the top K metrics for a metric name.
type Processor struct {
	metricName string
	topK       int
}

// Process filters metrics keeping only the top K for the configured metric name.
func (p *Processor) Process(metrics []otel.Metric) []otel.Metric {
	var target []otel.Metric
	var rest []otel.Metric
	for _, m := range metrics {
		if m.Name == p.metricName {
			target = append(target, m)
		} else {
			rest = append(rest, m)
		}
	}
	sort.Slice(target, func(i, j int) bool { return target[i].Value > target[j].Value })
	if len(target) > p.topK {
		target = target[:p.topK]
	}
	return append(target, rest...)
}

// Factory creates top-k processors.
type Factory struct{}

// Type returns the processor type.
func (Factory) Type() string { return "phoenix/topk" }

// Create instantiates a new processor from config.
func (Factory) Create(cfg map[string]interface{}) (otel.Processor, error) {
	metricName, _ := cfg["metric_name"].(string)
	if metricName == "" {
		return nil, fmt.Errorf("metric_name required")
	}
	var k int
	switch v := cfg["top_k"].(type) {
	case int:
		k = v
	case float64:
		k = int(v)
	}
	if k <= 0 {
		return nil, fmt.Errorf("top_k must be > 0")
	}
	return &Processor{metricName: metricName, topK: k}, nil
}

func init() {
	otel.RegisterProcessorFactory(&Factory{})
}
