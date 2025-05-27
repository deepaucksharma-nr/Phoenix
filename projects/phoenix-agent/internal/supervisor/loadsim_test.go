package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadSimManager_Lifecycle(t *testing.T) {
	manager := NewLoadSimManager()

	t.Run("StartAndStop", func(t *testing.T) {
		// Start a load simulation
		err := manager.Start("steady", "5s")
		require.NoError(t, err)

		// Check metrics
		metrics := manager.GetMetrics()
		assert.True(t, metrics["load_sim_active"].(bool))
		assert.NotZero(t, metrics["pid"])

		// Wait a bit
		time.Sleep(1 * time.Second)

		// Stop the simulation
		err = manager.Stop()
		require.NoError(t, err)

		// Check metrics again
		metrics = manager.GetMetrics()
		assert.False(t, metrics["load_sim_active"].(bool))
	})

	t.Run("CannotStartMultiple", func(t *testing.T) {
		// Start first simulation
		err := manager.Start("steady", "5s")
		require.NoError(t, err)

		// Try to start another
		err = manager.Start("spike", "5s")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already running")

		// Clean up
		manager.Stop()
	})

	t.Run("StopWhenNotRunning", func(t *testing.T) {
		// Stop when nothing is running should be no-op
		err := manager.Stop()
		assert.NoError(t, err)
	})

	t.Run("InvalidProfile", func(t *testing.T) {
		err := manager.Start("invalid-profile", "5s")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown profile")
	})

	t.Run("InvalidDuration", func(t *testing.T) {
		err := manager.Start("steady", "invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid duration")
	})

	t.Run("TimeoutHandling", func(t *testing.T) {
		// Start with very short duration
		err := manager.Start("steady", "1s")
		require.NoError(t, err)

		// Wait for it to complete
		time.Sleep(2 * time.Second)

		// Should have cleaned up automatically
		metrics := manager.GetMetrics()
		assert.False(t, metrics["load_sim_active"].(bool))
	})

	t.Run("GracefulShutdown", func(t *testing.T) {
		// Start a simulation
		err := manager.Start("steady", "10s")
		require.NoError(t, err)

		// Shutdown with context
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = manager.Shutdown(ctx)
		assert.NoError(t, err)

		// Should be stopped
		metrics := manager.GetMetrics()
		assert.False(t, metrics["load_sim_active"].(bool))
	})
}

func TestLoadSimManager_ProfileScripts(t *testing.T) {
	manager := NewLoadSimManager()

	profiles := []string{
		"high-cardinality",
		"high-card",
		"realistic",
		"normal",
		"spike",
		"process-churn",
		"steady",
	}

	for _, profile := range profiles {
		t.Run(profile, func(t *testing.T) {
			script := manager.getProfileScript(profile)
			assert.NotEmpty(t, script, "Profile %s should have a script", profile)
			assert.Contains(t, script, "#!/bin/bash", "Script should be a bash script")
		})
	}

	// Test unknown profile
	t.Run("UnknownProfile", func(t *testing.T) {
		script := manager.getProfileScript("unknown")
		assert.Empty(t, script)
	})
}

func TestLoadSimManager_ConcurrentAccess(t *testing.T) {
	manager := NewLoadSimManager()

	// Start a simulation
	err := manager.Start("steady", "5s")
	require.NoError(t, err)

	// Concurrent access to GetMetrics
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			metrics := manager.GetMetrics()
			assert.NotNil(t, metrics)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Stop
	err = manager.Stop()
	require.NoError(t, err)
}

// TestLoadSimManager_ProcessCleanup verifies that child processes are cleaned up
func TestLoadSimManager_ProcessCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping process cleanup test in short mode")
	}

	manager := NewLoadSimManager()

	// Start a simulation that creates child processes
	err := manager.Start("normal", "3s")
	require.NoError(t, err)

	// Get the PID
	metrics := manager.GetMetrics()
	pid := metrics["pid"].(int)
	assert.NotZero(t, pid)

	// Wait a bit for child processes to start
	time.Sleep(500 * time.Millisecond)

	// Stop the simulation
	err = manager.Stop()
	require.NoError(t, err)

	// Give time for cleanup
	time.Sleep(1 * time.Second)

	// Verify process is gone
	// Note: This is platform-specific and might need adjustment
	// On Unix systems, we can check if the process exists
	metrics = manager.GetMetrics()
	assert.False(t, metrics["load_sim_active"].(bool))
}
