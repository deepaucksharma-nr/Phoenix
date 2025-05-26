package adaptivefilter

import (
	"fmt"

	"github.com/phoenix/platform/pkg/otel"
)

// Config defines thresholds for filtering.
type Config struct {
	CPUPercent float64
	MemoryMB   float64
}

// Processor drops metrics below configured thresholds.
type Processor struct {
	cpu float64
	mem float64
}

// Process filters metrics based on CPU and memory thresholds.
func (p *Processor) Process(metrics []otel.Metric) []otel.Metric {
	var out []otel.Metric
	for _, m := range metrics {
		if m.Name == "process.cpu.utilization" && m.Value < p.cpu {
			continue
		}
		if m.Name == "process.memory.usage" && m.Value < p.mem {
			continue
		}
		out = append(out, m)
	}
	return out
}

// Factory creates adaptive filter processors.
type Factory struct{}

// Type returns the processor type name.
func (Factory) Type() string { return "phoenix/adaptive_filter" }

// Create instantiates a processor from config.
func (Factory) Create(cfg map[string]interface{}) (otel.Processor, error) {
	thresholds, ok := cfg["base_thresholds"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("base_thresholds required")
	}
	cpu := toFloat64(thresholds["cpu_percent"])
	mem := toFloat64(thresholds["memory_mb"])
	return &Processor{cpu: cpu, mem: mem}, nil
}

func toFloat64(v interface{}) float64 {
	switch t := v.(type) {
	case int:
		return float64(t)
	case float64:
		return t
	default:
		return 0
	}
}

func init() {
	otel.RegisterProcessorFactory(&Factory{})
}
