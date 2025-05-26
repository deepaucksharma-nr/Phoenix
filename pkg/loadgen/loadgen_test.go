package loadgen

import (
	"context"
	"testing"
	"time"
)

func TestMemoryProcessSpawner(t *testing.T) {
	spawner := NewMemoryProcessSpawner()

	// Test spawning a process
	config := ProcessConfig{
		Name:      "test-process",
		CPUTarget: 25.0,
		MemoryMB:  512,
		Duration:  100 * time.Millisecond,
		Tags: map[string]string{
			"test": "true",
		},
	}

	proc, err := spawner.SpawnProcess(config)
	if err != nil {
		t.Fatalf("Failed to spawn process: %v", err)
	}

	if proc.Name != "test-process" {
		t.Errorf("Expected process name 'test-process', got '%s'", proc.Name)
	}

	// Test listing processes
	processes, err := spawner.ListProcesses()
	if err != nil {
		t.Fatalf("Failed to list processes: %v", err)
	}

	if len(processes) != 1 {
		t.Errorf("Expected 1 process, got %d", len(processes))
	}

	// Test updating process
	err = spawner.UpdateProcess(proc.PID, 50.0, 1024)
	if err != nil {
		t.Fatalf("Failed to update process: %v", err)
	}

	// Test killing process
	err = spawner.KillProcess(proc.PID)
	if err != nil {
		t.Fatalf("Failed to kill process: %v", err)
	}

	// Wait for cleanup
	time.Sleep(200 * time.Millisecond)

	processes, err = spawner.ListProcesses()
	if err != nil {
		t.Fatalf("Failed to list processes after kill: %v", err)
	}

	if len(processes) != 0 {
		t.Errorf("Expected 0 processes after kill, got %d", len(processes))
	}
}

func TestLoadPatternFactory(t *testing.T) {
	spawner := NewMemoryProcessSpawner()
	factory := NewLoadPatternFactory(spawner)

	// Test getting available profiles
	profiles := factory.GetAvailableProfiles()
	if len(profiles) < 3 {
		t.Errorf("Expected at least 3 profiles, got %d", len(profiles))
	}

	// Test creating realistic pattern
	pattern, err := factory.CreateLoadPattern(LoadPatternRealistic, nil)
	if err != nil {
		t.Fatalf("Failed to create realistic pattern: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Run pattern briefly
	go pattern.Generate(ctx)

	// Wait for some processes to be created
	time.Sleep(500 * time.Millisecond)

	// Check metrics
	metrics := pattern.GetMetrics()
	if metrics.ProcessCount == 0 {
		t.Error("Expected some processes to be created")
	}

	// Stop pattern
	err = pattern.Stop()
	if err != nil {
		t.Errorf("Failed to stop pattern: %v", err)
	}
}

func TestDistributionGeneration(t *testing.T) {
	tests := []struct {
		name string
		dist Distribution
	}{
		{
			name: "uniform",
			dist: Distribution{
				Type: "uniform",
				Min:  10,
				Max:  20,
			},
		},
		{
			name: "normal",
			dist: Distribution{
				Type:   "normal",
				Min:    0,
				Max:    100,
				Mean:   50,
				StdDev: 10,
			},
		},
		{
			name: "exponential",
			dist: Distribution{
				Type: "exponential",
				Min:  0,
				Max:  100,
				Mean: 20,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate multiple values
			for i := 0; i < 100; i++ {
				value := generateValue(tt.dist)
				if value < tt.dist.Min || value > tt.dist.Max {
					t.Errorf("Generated value %f outside range [%f, %f]", 
						value, tt.dist.Min, tt.dist.Max)
				}
			}
		})
	}
}