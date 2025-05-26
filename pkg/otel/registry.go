package otel

import "github.com/phoenix/platform/pkg/analytics"

// Metric represents a single metric data point used by processors.
type Metric = analytics.Metric

// Processor processes a slice of metrics and returns the result.
type Processor interface {
	Process([]Metric) []Metric
}

// ProcessorFactory creates processors based on configuration.
type ProcessorFactory interface {
	// Type returns the unique processor type name.
	Type() string
	// Create creates a processor from config options.
	Create(config map[string]interface{}) (Processor, error)
}

var processorFactories = map[string]ProcessorFactory{}

// RegisterProcessorFactory registers a processor factory.
func RegisterProcessorFactory(factory ProcessorFactory) {
	if factory == nil {
		return
	}
	processorFactories[factory.Type()] = factory
}

// GetProcessorFactory retrieves a registered factory by name.
func GetProcessorFactory(name string) (ProcessorFactory, bool) {
	f, ok := processorFactories[name]
	return f, ok
}

// ClearProcessorFactories removes all registered factories.
func ClearProcessorFactories() {
	processorFactories = map[string]ProcessorFactory{}
}
