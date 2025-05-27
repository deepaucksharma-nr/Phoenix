package supervisor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type LoadSimManager struct {
	activeJob   *exec.Cmd
	activeJobMu sync.Mutex
	cancelFunc  context.CancelFunc
	cleanupChan chan struct{}
	cleanupWg   sync.WaitGroup
}

func NewLoadSimManager() *LoadSimManager {
	return &LoadSimManager{
		cleanupChan: make(chan struct{}),
	}
}

// Start starts a load simulation with the given profile
func (m *LoadSimManager) Start(profile, durationStr string) error {
	m.activeJobMu.Lock()
	defer m.activeJobMu.Unlock()

	if m.activeJob != nil {
		return fmt.Errorf("load simulation already running")
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	// Validate profile and duration
	if err := m.ValidateProfile(profile, duration); err != nil {
		return err
	}

	script := m.getProfileScript(profile)
	if script == "" {
		return fmt.Errorf("unknown profile: %s", profile)
	}

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	m.cancelFunc = cancel

	// Set environment variables for the script
	env := os.Environ()
	env = append(env, fmt.Sprintf("LOAD_PROFILE=%s", profile))
	env = append(env, fmt.Sprintf("LOAD_DURATION=%s", durationStr))

	m.activeJob = exec.CommandContext(ctx, "bash", "-c", script)
	m.activeJob.Env = env

	// Start the job
	if err := m.activeJob.Start(); err != nil {
		m.activeJob = nil
		m.cancelFunc = nil
		return fmt.Errorf("failed to start load simulation: %w", err)
	}

	// Monitor job in background
	m.cleanupWg.Add(1)
	go m.monitorJob(profile, duration)

	// Log profile information
	var profileInfo ProfileInfo
	for _, info := range m.GetAvailableProfiles() {
		if info.Name == profile || m.isAlias(profile, info.Name) {
			profileInfo = info
			break
		}
	}

	log.Info().
		Str("profile", profile).
		Str("description", profileInfo.Description).
		Dur("duration", duration).
		Int("pid", m.activeJob.Process.Pid).
		Str("cpu_impact", profileInfo.ResourceUsage.CPU).
		Str("memory_impact", profileInfo.ResourceUsage.Memory).
		Str("network_impact", profileInfo.ResourceUsage.Network).
		Msg("Started load simulation")

	return nil
}

// Stop stops the current load simulation
func (m *LoadSimManager) Stop() error {
	m.activeJobMu.Lock()
	defer m.activeJobMu.Unlock()

	if m.activeJob == nil {
		return nil // Nothing to stop
	}

	pid := 0
	if m.activeJob.Process != nil {
		pid = m.activeJob.Process.Pid
	}

	// Cancel context to stop the job
	if m.cancelFunc != nil {
		m.cancelFunc()
	}

	// Send SIGTERM first for graceful shutdown
	if m.activeJob.Process != nil {
		m.activeJob.Process.Signal(os.Interrupt)

		// Give it 2 seconds to terminate gracefully
		done := make(chan struct{})
		go func() {
			m.activeJob.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Process terminated gracefully
		case <-time.After(2 * time.Second):
			// Force kill if not terminated
			m.activeJob.Process.Kill()
		}
	}

	// Wait for monitor goroutine to finish
	m.cleanupWg.Wait()

	m.activeJob = nil
	m.cancelFunc = nil

	log.Info().Int("pid", pid).Msg("Stopped load simulation")
	return nil
}

// GetMetrics returns metrics about the load simulation
func (m *LoadSimManager) GetMetrics() map[string]interface{} {
	m.activeJobMu.Lock()
	defer m.activeJobMu.Unlock()

	if m.activeJob == nil {
		return map[string]interface{}{
			"load_sim_active": false,
		}
	}

	pid := 0
	if m.activeJob.Process != nil {
		pid = m.activeJob.Process.Pid
	}

	return map[string]interface{}{
		"load_sim_active": true,
		"pid":             pid,
	}
}

// Shutdown gracefully shuts down the load simulation manager
func (m *LoadSimManager) Shutdown(ctx context.Context) error {
	log.Info().Msg("Shutting down load simulation manager")

	// Stop any active load simulation
	if err := m.Stop(); err != nil {
		log.Error().Err(err).Msg("Error stopping load simulation during shutdown")
	}

	// Close cleanup channel
	close(m.cleanupChan)

	// Wait for all goroutines to finish or context to expire
	done := make(chan struct{})
	go func() {
		m.cleanupWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info().Msg("Load simulation manager shutdown complete")
		return nil
	case <-ctx.Done():
		log.Warn().Msg("Load simulation manager shutdown timed out")
		return ctx.Err()
	}
}

func (m *LoadSimManager) monitorJob(profile string, duration time.Duration) {
	defer m.cleanupWg.Done()

	// Create a timer for maximum duration
	timer := time.NewTimer(duration + 30*time.Second) // Extra 30s buffer
	defer timer.Stop()

	pid := 0
	if m.activeJob.Process != nil {
		pid = m.activeJob.Process.Pid
	}

	// Monitor the job
	jobDone := make(chan error, 1)
	go func() {
		jobDone <- m.activeJob.Wait()
	}()

	var err error
	select {
	case err = <-jobDone:
		// Job completed
	case <-timer.C:
		// Timeout - force kill
		log.Warn().
			Str("profile", profile).
			Int("pid", pid).
			Dur("duration", duration).
			Msg("Load simulation exceeded maximum duration, forcing termination")

		if m.activeJob.Process != nil {
			m.activeJob.Process.Kill()
		}
		err = <-jobDone
	}

	// Cleanup any child processes
	m.cleanupChildProcesses(pid)

	m.activeJobMu.Lock()
	m.activeJob = nil
	m.cancelFunc = nil
	m.activeJobMu.Unlock()

	if err != nil {
		// Check if it's a context cancellation or signal termination
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == -1 || exitErr.ExitCode() == 143 { // -1 or SIGTERM
				log.Info().
					Str("profile", profile).
					Int("pid", pid).
					Msg("Load simulation was terminated")
			} else {
				log.Error().
					Err(err).
					Str("profile", profile).
					Int("pid", pid).
					Int("exit_code", exitErr.ExitCode()).
					Msg("Load simulation ended with error")
			}
		} else {
			log.Error().
				Err(err).
				Str("profile", profile).
				Int("pid", pid).
				Msg("Load simulation error")
		}
	} else {
		log.Info().
			Str("profile", profile).
			Int("pid", pid).
			Msg("Load simulation completed successfully")
	}
}

// cleanupChildProcesses kills any orphaned child processes
func (m *LoadSimManager) cleanupChildProcesses(parentPID int) {
	if parentPID == 0 {
		return
	}

	// Use pkill to clean up any child processes
	cmd := exec.Command("pkill", "-P", fmt.Sprintf("%d", parentPID))
	if err := cmd.Run(); err != nil {
		// It's okay if this fails - processes might already be gone
		log.Debug().
			Int("parent_pid", parentPID).
			Err(err).
			Msg("Failed to cleanup child processes")
	}
}

// ProfileInfo contains metadata about a load profile
type ProfileInfo struct {
	Name          string
	Description   string
	MaxDuration   time.Duration
	ResourceUsage struct {
		CPU     string // low, medium, high, variable
		Memory  string // minimal, low, medium, high
		Network string // minimal, low, medium, high, variable
	}
}

// GetAvailableProfiles returns information about all available profiles
func (m *LoadSimManager) GetAvailableProfiles() []ProfileInfo {
	return []ProfileInfo{
		{
			Name:        "high-cardinality",
			Description: "Simulates high cardinality metrics explosion",
			MaxDuration: 30 * time.Minute,
			ResourceUsage: struct {
				CPU     string
				Memory  string
				Network string
			}{
				CPU:     "low",
				Memory:  "high",
				Network: "medium",
			},
		},
		{
			Name:        "realistic",
			Description: "Simulates normal production workload",
			MaxDuration: time.Hour,
			ResourceUsage: struct {
				CPU     string
				Memory  string
				Network string
			}{
				CPU:     "medium",
				Memory:  "low",
				Network: "low",
			},
		},
		{
			Name:        "spike",
			Description: "Simulates traffic spikes and recovery",
			MaxDuration: 10 * time.Minute,
			ResourceUsage: struct {
				CPU     string
				Memory  string
				Network string
			}{
				CPU:     "variable",
				Memory:  "low",
				Network: "variable",
			},
		},
		{
			Name:        "steady",
			Description: "Maintains constant load for stability testing",
			MaxDuration: 24 * time.Hour,
			ResourceUsage: struct {
				CPU     string
				Memory  string
				Network string
			}{
				CPU:     "low",
				Memory:  "minimal",
				Network: "low",
			},
		},
	}
}

// ValidateProfile checks if a profile exists and validates the duration
func (m *LoadSimManager) ValidateProfile(profile string, duration time.Duration) error {
	// Check if profile exists
	script := m.getProfileScript(profile)
	if script == "" {
		return fmt.Errorf("unknown profile: %s", profile)
	}

	// Get profile info
	var maxDuration time.Duration
	for _, info := range m.GetAvailableProfiles() {
		if info.Name == profile || m.isAlias(profile, info.Name) {
			maxDuration = info.MaxDuration
			break
		}
	}

	// Validate duration
	if duration < time.Second {
		return fmt.Errorf("duration must be at least 1 second")
	}

	if maxDuration > 0 && duration > maxDuration {
		return fmt.Errorf("duration %v exceeds maximum %v for profile %s", duration, maxDuration, profile)
	}

	return nil
}

// isAlias checks if the given name is an alias for a profile
func (m *LoadSimManager) isAlias(alias, profile string) bool {
	aliases := map[string][]string{
		"high-cardinality": {"high-card"},
		"realistic":        {"normal"},
		"steady":           {"process-churn"},
	}

	if profileAliases, ok := aliases[profile]; ok {
		for _, a := range profileAliases {
			if a == alias {
				return true
			}
		}
	}

	return false
}

func (m *LoadSimManager) getProfileScript(profile string) string {
	// Map CLI profile names to internal profile names
	switch profile {
	case "high-cardinality", "high-card":
		return m.getHighCardinalityScript()
	case "realistic", "normal":
		return m.getNormalLoadScript()
	case "spike":
		return m.getSpikeLoadScript()
	case "process-churn", "steady":
		return m.getSteadyLoadScript()
	case "custom":
		// TODO: Implement custom profile support
		return ""
	default:
		return ""
	}
}

func (m *LoadSimManager) getHighCardinalityScript() string {
	return `
#!/bin/bash
# High cardinality metrics generator
ENDPOINT="${OTEL_ENDPOINT:-http://localhost:4318}"

while true; do
    # Generate metrics with high cardinality
    for i in {1..1000}; do
        # Generate unique user ID
        USER_ID="user-$(uuidgen | cut -d'-' -f1)"
        
        # Send metric via OTLP HTTP
        curl -X POST "${ENDPOINT}/v1/metrics" \
            -H "Content-Type: application/json" \
            -d "{
                \"resourceMetrics\": [{
                    \"resource\": {
                        \"attributes\": [{
                            \"key\": \"service.name\",
                            \"value\": {\"stringValue\": \"load-test\"}
                        }]
                    },
                    \"scopeMetrics\": [{
                        \"metrics\": [{
                            \"name\": \"http.request.duration\",
                            \"unit\": \"ms\",
                            \"histogram\": {
                                \"dataPoints\": [{
                                    \"startTimeUnixNano\": \"$(date +%s)000000000\",
                                    \"timeUnixNano\": \"$(date +%s)000000000\",
                                    \"count\": \"1\",
                                    \"sum\": $((RANDOM % 1000)),
                                    \"bucketCounts\": [0, 1, 0, 0, 0],
                                    \"explicitBounds\": [10, 50, 100, 500, 1000],
                                    \"attributes\": [
                                        {
                                            \"key\": \"user.id\",
                                            \"value\": {\"stringValue\": \"${USER_ID}\"}
                                        },
                                        {
                                            \"key\": \"endpoint\",
                                            \"value\": {\"stringValue\": \"/api/v1/endpoint-${i}\"}
                                        },
                                        {
                                            \"key\": \"status_code\",
                                            \"value\": {\"intValue\": \"200\"}
                                        }
                                    ]
                                }]
                            }
                        }]
                    }]
                }]
            }" 2>/dev/null
    done
    sleep 1
done
`
}

func (m *LoadSimManager) getNormalLoadScript() string {
	return `
#!/bin/bash
# Normal load using stress-ng
if command -v stress-ng &> /dev/null; then
    stress-ng --cpu 2 --io 2 --vm 1 --vm-bytes 128M --timeout 60s
else
    # Fallback to simple CPU load
    for i in {1..4}; do
        while true; do
            echo "scale=10000; 4*a(1)" | bc -l > /dev/null
        done &
    done
    
    # Let it run for duration
    sleep 60
    
    # Kill all background jobs
    jobs -p | xargs kill 2>/dev/null
fi
`
}

func (m *LoadSimManager) getSpikeLoadScript() string {
	return `
#!/bin/bash
# Spike load pattern
ENDPOINT="${OTEL_ENDPOINT:-http://localhost:4318}"

# Normal load for 30s
for i in {1..30}; do
    curl -X POST "${ENDPOINT}/v1/metrics" \
        -H "Content-Type: application/json" \
        -d "{
            \"resourceMetrics\": [{
                \"resource\": {
                    \"attributes\": [{
                        \"key\": \"service.name\",
                        \"value\": {\"stringValue\": \"spike-test\"}
                    }]
                },
                \"scopeMetrics\": [{
                    \"metrics\": [{
                        \"name\": \"system.load\",
                        \"gauge\": {
                            \"dataPoints\": [{
                                \"timeUnixNano\": \"$(date +%s)000000000\",
                                \"asDouble\": 1.5
                            }]
                        }
                    }]
                }]
            }]
        }" 2>/dev/null
    sleep 1
done

# Spike for 10s
for i in {1..10}; do
    # Send 100x normal rate
    for j in {1..100}; do
        curl -X POST "${ENDPOINT}/v1/metrics" \
            -H "Content-Type: application/json" \
            -d "{
                \"resourceMetrics\": [{
                    \"resource\": {
                        \"attributes\": [{
                            \"key\": \"service.name\",
                            \"value\": {\"stringValue\": \"spike-test\"}
                        }]
                    },
                    \"scopeMetrics\": [{
                        \"metrics\": [{
                            \"name\": \"system.load\",
                            \"gauge\": {
                                \"dataPoints\": [{
                                    \"timeUnixNano\": \"$(date +%s)000000000\",
                                    \"asDouble\": 15.0
                                }]
                            }
                        }]
                    }]
                }]
            }" 2>/dev/null &
    done
    wait
    sleep 1
done

# Return to normal
for i in {1..20}; do
    curl -X POST "${ENDPOINT}/v1/metrics" \
        -H "Content-Type: application/json" \
        -d "{
            \"resourceMetrics\": [{
                \"resource\": {
                    \"attributes\": [{
                        \"key\": \"service.name\",
                        \"value\": {\"stringValue\": \"spike-test\"}
                    }]
                },
                \"scopeMetrics\": [{
                    \"metrics\": [{
                        \"name\": \"system.load\",
                        \"gauge\": {
                            \"dataPoints\": [{
                                \"timeUnixNano\": \"$(date +%s)000000000\",
                                \"asDouble\": 1.5
                            }]
                        }
                    }]
                }]
            }]
        }" 2>/dev/null
    sleep 1
done
`
}

func (m *LoadSimManager) getSteadyLoadScript() string {
	return `
#!/bin/bash
# Steady load pattern
ENDPOINT="${OTEL_ENDPOINT:-http://localhost:4318}"
RATE=10 # requests per second

while true; do
    for i in $(seq 1 $RATE); do
        curl -X POST "${ENDPOINT}/v1/metrics" \
            -H "Content-Type: application/json" \
            -d "{
                \"resourceMetrics\": [{
                    \"resource\": {
                        \"attributes\": [{
                            \"key\": \"service.name\",
                            \"value\": {\"stringValue\": \"steady-load\"}
                        }]
                    },
                    \"scopeMetrics\": [{
                        \"metrics\": [{
                            \"name\": \"http.requests\",
                            \"sum\": {
                                \"dataPoints\": [{
                                    \"startTimeUnixNano\": \"$(date +%s)000000000\",
                                    \"timeUnixNano\": \"$(date +%s)000000000\",
                                    \"asDouble\": 1,
                                    \"attributes\": [{
                                        \"key\": \"method\",
                                        \"value\": {\"stringValue\": \"GET\"}
                                    }]
                                }],
                                \"aggregationTemporality\": 2,
                                \"isMonotonic\": true
                            }
                        }]
                    }]
                }]
            }" 2>/dev/null &
    done
    
    # Sleep to maintain rate
    sleep 1
done
`
}
