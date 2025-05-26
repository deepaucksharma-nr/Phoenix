package loadgen

import (
	"fmt"
	"time"
)

// LoadPatternType represents the type of load pattern
type LoadPatternType string

const (
	LoadPatternRealistic      LoadPatternType = "realistic"
	LoadPatternHighCardinality LoadPatternType = "high-cardinality"
	LoadPatternProcessChurn   LoadPatternType = "process-churn"
	LoadPatternCustom         LoadPatternType = "custom"
)

// DefaultProfiles provides pre-configured load profiles
var DefaultProfiles = map[LoadPatternType]ProfileConfig{
	LoadPatternRealistic: {
		Name:             "realistic",
		ProcessCount:     100,
		ProcessChurnRate: 0.1, // 10% churn per second
		CPUDistribution: Distribution{
			Type: "normal",
			Min:  0,
			Max:  80,
			Mean: 20,
			StdDev: 15,
		},
		MemoryDistribution: Distribution{
			Type: "normal",
			Min:  50,
			Max:  2048,
			Mean: 256,
			StdDev: 200,
		},
		DurationRange: DurationRange{
			Min: 30 * time.Second,
			Max: 5 * time.Minute,
		},
		Tags: map[string]string{
			"profile": "realistic",
		},
	},
	LoadPatternHighCardinality: {
		Name:             "high-cardinality",
		ProcessCount:     500,
		ProcessChurnRate: 0.5, // 50% churn per second
		CPUDistribution: Distribution{
			Type: "uniform",
			Min:  5,
			Max:  50,
		},
		MemoryDistribution: Distribution{
			Type: "uniform",
			Min:  100,
			Max:  500,
		},
		DurationRange: DurationRange{
			Min: 10 * time.Second,
			Max: 60 * time.Second,
		},
		Tags: map[string]string{
			"profile": "high-cardinality",
		},
	},
	LoadPatternProcessChurn: {
		Name:             "process-churn",
		ProcessCount:     200,
		ProcessChurnRate: 2.0, // 200% churn per second
		CPUDistribution: Distribution{
			Type: "exponential",
			Min:  1,
			Max:  100,
			Mean: 10,
		},
		MemoryDistribution: Distribution{
			Type: "exponential",
			Min:  50,
			Max:  1024,
			Mean: 128,
		},
		DurationRange: DurationRange{
			Min: 1 * time.Second,
			Max: 10 * time.Second,
		},
		Tags: map[string]string{
			"profile": "process-churn",
		},
	},
}

// LoadPatternFactory creates load patterns
type LoadPatternFactory struct {
	spawner ProcessSpawner
}

// NewLoadPatternFactory creates a new load pattern factory
func NewLoadPatternFactory(spawner ProcessSpawner) *LoadPatternFactory {
	return &LoadPatternFactory{
		spawner: spawner,
	}
}

// CreateLoadPattern creates a load pattern based on type
func (f *LoadPatternFactory) CreateLoadPattern(patternType LoadPatternType, config *ProfileConfig) (LoadPattern, error) {
	// Use default config if not provided
	if config == nil {
		defaultConfig, ok := DefaultProfiles[patternType]
		if !ok {
			return nil, fmt.Errorf("unknown load pattern type: %s", patternType)
		}
		config = &defaultConfig
	}

	switch patternType {
	case LoadPatternRealistic:
		return NewRealisticLoadPattern(f.spawner, *config), nil
	case LoadPatternHighCardinality:
		return NewHighCardinalityLoadPattern(f.spawner, *config), nil
	case LoadPatternProcessChurn:
		return NewProcessChurnLoadPattern(f.spawner, *config), nil
	case LoadPatternCustom:
		// For custom patterns, config must be provided
		if config == nil {
			return nil, fmt.Errorf("custom pattern requires configuration")
		}
		// TODO: Implement custom pattern interpreter
		return nil, fmt.Errorf("custom patterns not yet implemented")
	default:
		return nil, fmt.Errorf("unknown load pattern type: %s", patternType)
	}
}

// GetAvailableProfiles returns all available profile names
func (f *LoadPatternFactory) GetAvailableProfiles() []string {
	profiles := make([]string, 0, len(DefaultProfiles))
	for profile := range DefaultProfiles {
		profiles = append(profiles, string(profile))
	}
	return profiles
}

// GetProfileConfig returns the configuration for a profile
func (f *LoadPatternFactory) GetProfileConfig(patternType LoadPatternType) (*ProfileConfig, error) {
	config, ok := DefaultProfiles[patternType]
	if !ok {
		return nil, fmt.Errorf("unknown load pattern type: %s", patternType)
	}
	return &config, nil
}