package supervisor

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type LoadSimManager struct {
	activeJob   *exec.Cmd
	activeJobMu sync.Mutex
	cancelFunc  context.CancelFunc
}

func NewLoadSimManager() *LoadSimManager {
	return &LoadSimManager{}
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

	script := m.getProfileScript(profile)
	if script == "" {
		return fmt.Errorf("unknown profile: %s", profile)
	}

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	m.cancelFunc = cancel

	m.activeJob = exec.CommandContext(ctx, "bash", "-c", script)

	// Start the job
	if err := m.activeJob.Start(); err != nil {
		m.activeJob = nil
		m.cancelFunc = nil
		return fmt.Errorf("failed to start load simulation: %w", err)
	}

	// Monitor job in background
	go m.monitorJob(profile)

	log.Info().
		Str("profile", profile).
		Dur("duration", duration).
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

	// Cancel context to stop the job
	if m.cancelFunc != nil {
		m.cancelFunc()
	}

	// Also send interrupt signal
	if m.activeJob.Process != nil {
		m.activeJob.Process.Kill()
	}

	m.activeJob = nil
	m.cancelFunc = nil

	log.Info().Msg("Stopped load simulation")
	return nil
}

// GetMetrics returns metrics about the load simulation
func (m *LoadSimManager) GetMetrics() map[string]interface{} {
	m.activeJobMu.Lock()
	defer m.activeJobMu.Unlock()

	if m.activeJob == nil {
		return nil
	}

	return map[string]interface{}{
		"load_sim_active": true,
		"pid":             m.activeJob.Process.Pid,
	}
}

func (m *LoadSimManager) monitorJob(profile string) {
	err := m.activeJob.Wait()

	m.activeJobMu.Lock()
	m.activeJob = nil
	m.cancelFunc = nil
	m.activeJobMu.Unlock()

	if err != nil {
		// Check if it's a context cancellation
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == -1 {
			log.Info().Str("profile", profile).Msg("Load simulation was cancelled")
		} else {
			log.Error().Err(err).Str("profile", profile).Msg("Load simulation ended with error")
		}
	} else {
		log.Info().Str("profile", profile).Msg("Load simulation completed successfully")
	}
}

func (m *LoadSimManager) getProfileScript(profile string) string {
	switch profile {
	case "high-card":
		return m.getHighCardinalityScript()
	case "normal":
		return m.getNormalLoadScript()
	case "spike":
		return m.getSpikeLoadScript()
	case "steady":
		return m.getSteadyLoadScript()
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