package generator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewLoadGenerator(t *testing.T) {
	logger := zaptest.NewLogger(t)
	profile := "realistic"
	processCount := 100
	duration := 30 * time.Minute

	lg := NewLoadGenerator(logger, profile, processCount, duration)

	assert.NotNil(t, lg)
	assert.Equal(t, profile, lg.profile)
	assert.Equal(t, processCount, lg.processCount)
	assert.Equal(t, duration, lg.duration)
	assert.NotNil(t, lg.processes)
	assert.NotNil(t, lg.ctx)
	assert.NotNil(t, lg.cancel)
}

func TestGetCPULoad(t *testing.T) {
	logger := zaptest.NewLogger(t)
	lg := NewLoadGenerator(logger, "test", 10, time.Minute)

	tests := []struct {
		pattern string
		minLoad float64
		maxLoad float64
	}{
		{"steady", 0.3, 0.4},
		{"spiky", 0.1, 1.0}, // Can be either low or high
		{"growing", 0.1, 0.7},
		{"random", 0.0, 1.0},
		{"unknown", 0.3, 0.3}, // Default case
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			load := lg.getCPULoad(tt.pattern)
			assert.GreaterOrEqual(t, load, tt.minLoad, "CPU load too low for pattern %s", tt.pattern)
			assert.LessOrEqual(t, load, tt.maxLoad, "CPU load too high for pattern %s", tt.pattern)
		})
	}
}

func TestGetMemLoad(t *testing.T) {
	logger := zaptest.NewLogger(t)
	lg := NewLoadGenerator(logger, "test", 10, time.Minute)

	tests := []struct {
		pattern string
		minLoad float64
		maxLoad float64
	}{
		{"steady", 0.2, 0.3},
		{"spiky", 0.1, 1.0}, // Can be either low or high
		{"growing", 0.1, 0.8},
		{"random", 0.0, 1.0},
		{"unknown", 0.2, 0.2}, // Default case
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			load := lg.getMemLoad(tt.pattern)
			assert.GreaterOrEqual(t, load, tt.minLoad, "Memory load too low for pattern %s", tt.pattern)
			assert.LessOrEqual(t, load, tt.maxLoad, "Memory load too high for pattern %s", tt.pattern)
		})
	}
}

func TestGetActiveProcessCount(t *testing.T) {
	logger := zaptest.NewLogger(t)
	lg := NewLoadGenerator(logger, "test", 10, time.Minute)

	// Initially should be 0
	count := lg.GetActiveProcessCount()
	assert.Equal(t, 0, count)

	// Add some mock processes
	lg.processes["test1"] = &Process{Name: "test1"}
	lg.processes["test2"] = &Process{Name: "test2"}

	count = lg.GetActiveProcessCount()
	assert.Equal(t, 2, count)
}

func TestGetProcessList(t *testing.T) {
	logger := zaptest.NewLogger(t)
	lg := NewLoadGenerator(logger, "test", 10, time.Minute)

	// Initially should be empty
	list := lg.GetProcessList()
	assert.Len(t, list, 0)

	// Add some mock processes
	lg.processes["test1"] = &Process{Name: "test1"}
	lg.processes["test2"] = &Process{Name: "test2"}
	lg.processes["test3"] = &Process{Name: "test3"}

	list = lg.GetProcessList()
	assert.Len(t, list, 3)
	assert.Contains(t, list, "test1")
	assert.Contains(t, list, "test2")
	assert.Contains(t, list, "test3")
}

func TestStartUnknownProfile(t *testing.T) {
	logger := zaptest.NewLogger(t)
	lg := NewLoadGenerator(logger, "unknown-profile", 10, time.Second)

	err := lg.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown profile")
}

func TestStop(t *testing.T) {
	logger := zaptest.NewLogger(t)
	lg := NewLoadGenerator(logger, "test", 10, time.Minute)

	// Should not error even if no processes are running
	err := lg.Stop()
	assert.NoError(t, err)

	// Context should be cancelled
	select {
	case <-lg.ctx.Done():
		// Expected
	default:
		t.Error("Context should be cancelled after Stop()")
	}
}

// This test is commented out because it would actually spawn processes
// Uncomment and run manually if you want to test process creation
/*
func TestCreateProcessIntegration(t *testing.T) {
	logger := zaptest.NewLogger(t)
	lg := NewLoadGenerator(logger, "test", 10, 5*time.Second)

	// This would create an actual process - only run in integration tests
	lg.wg.Add(1)
	go lg.createProcess("test-process", 2*time.Second, "steady", "steady")

	// Wait a bit
	time.Sleep(500 * time.Millisecond)

	// Check that process was added
	count := lg.GetActiveProcessCount()
	assert.Equal(t, 1, count)

	// Wait for process to finish
	lg.wg.Wait()

	// Process should be cleaned up
	count = lg.GetActiveProcessCount()
	assert.Equal(t, 0, count)
}
*/

func BenchmarkGetCPULoad(b *testing.B) {
	logger := zaptest.NewLogger(b)
	lg := NewLoadGenerator(logger, "test", 10, time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lg.getCPULoad("steady")
	}
}

func BenchmarkGetMemLoad(b *testing.B) {
	logger := zaptest.NewLogger(b)
	lg := NewLoadGenerator(logger, "test", 10, time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lg.getMemLoad("steady")
	}
}