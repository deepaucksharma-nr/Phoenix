package supervisor

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/phoenix/platform/projects/phoenix-agent/internal/config"
	"github.com/rs/zerolog/log"
)

type CollectorManager struct {
	config    *config.Config
	processes map[string]*Process
	mu        sync.RWMutex
}

type Process struct {
	ID      string
	Variant string
	Cmd     *exec.Cmd
	Pid     int
}

func NewCollectorManager(cfg *config.Config) *CollectorManager {
	return &CollectorManager{
		config:    cfg,
		processes: make(map[string]*Process),
	}
}

// Start starts a new OTel collector process
func (m *CollectorManager) Start(id, variant, configURL string, vars map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already running
	if _, exists := m.processes[id]; exists {
		return fmt.Errorf("collector %s already running", id)
	}

	// Download and process config
	config, err := m.downloadConfig(configURL)
	if err != nil {
		return fmt.Errorf("failed to download config: %w", err)
	}

	// Apply variable substitution
	processedConfig, err := m.applyVariables(config, vars, id, variant)
	if err != nil {
		return fmt.Errorf("failed to apply variables: %w", err)
	}

	// Write config to disk
	configPath := filepath.Join(m.config.ConfigDir, fmt.Sprintf("%s.yaml", id))
	if err := os.MkdirAll(m.config.ConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configPath, []byte(processedConfig), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Prepare command
	cmd := exec.Command(
		"otelcol-contrib",
		"--config", configPath,
		"--set", "service.telemetry.metrics.address=:0", // Disable default metrics endpoint
	)

	// Set environment variables
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("EXPERIMENT_ID=%s", strings.Split(id, "-")[0]),
		fmt.Sprintf("VARIANT=%s", variant),
		fmt.Sprintf("HOST_ID=%s", m.config.HostID),
	)

	// Set up logging
	logFile, err := os.Create(filepath.Join(m.config.ConfigDir, fmt.Sprintf("%s.log", id)))
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// Start process
	if err := cmd.Start(); err != nil {
		logFile.Close()
		return fmt.Errorf("failed to start collector: %w", err)
	}

	process := &Process{
		ID:      id,
		Variant: variant,
		Cmd:     cmd,
		Pid:     cmd.Process.Pid,
	}

	m.processes[id] = process

	// Monitor process in background
	go m.monitorProcess(process, logFile)

	log.Info().
		Str("id", id).
		Str("variant", variant).
		Int("pid", process.Pid).
		Msg("Started OTel collector")

	return nil
}

// Stop stops a collector process
func (m *CollectorManager) Stop(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	process, exists := m.processes[id]
	if !exists {
		return fmt.Errorf("collector %s not found", id)
	}

	// Send graceful shutdown signal
	if err := process.Cmd.Process.Signal(os.Interrupt); err != nil {
		log.Warn().Err(err).Str("id", id).Msg("Failed to send interrupt signal, killing process")
		process.Cmd.Process.Kill()
	}

	// Wait for process to exit (with timeout)
	done := make(chan error, 1)
	go func() {
		done <- process.Cmd.Wait()
	}()

	select {
	case <-done:
		log.Info().Str("id", id).Msg("Collector stopped gracefully")
	case <-time.After(10 * time.Second):
		log.Warn().Str("id", id).Msg("Collector stop timeout, force killing")
		process.Cmd.Process.Kill()
	}

	delete(m.processes, id)

	// Clean up config file
	configPath := filepath.Join(m.config.ConfigDir, fmt.Sprintf("%s.yaml", id))
	os.Remove(configPath)

	return nil
}

// StopAll stops all running collectors
func (m *CollectorManager) StopAll() {
	m.mu.Lock()
	ids := make([]string, 0, len(m.processes))
	for id := range m.processes {
		ids = append(ids, id)
	}
	m.mu.Unlock()

	for _, id := range ids {
		if err := m.Stop(id); err != nil {
			log.Error().Err(err).Str("id", id).Msg("Failed to stop collector")
		}
	}
}

// GetProcessInfo returns information about a running process
func (m *CollectorManager) GetProcessInfo(id string) map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	process, exists := m.processes[id]
	if !exists {
		return nil
	}

	return map[string]interface{}{
		"pid":     process.Pid,
		"variant": process.Variant,
	}
}

// GetMetrics returns metrics for all running collectors
func (m *CollectorManager) GetMetrics() []map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var metrics []map[string]interface{}
	for id, process := range m.processes {
		metrics = append(metrics, map[string]interface{}{
			"collector_id": id,
			"variant":      process.Variant,
			"pid":          process.Pid,
			"running":      true,
		})
	}

	return metrics
}

func (m *CollectorManager) downloadConfig(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read config: %w", err)
	}

	return string(data), nil
}

func (m *CollectorManager) applyVariables(config string, vars map[string]string, id, variant string) (string, error) {
	// Add default variables
	allVars := map[string]string{
		"EXPERIMENT_ID":          strings.Split(id, "-")[0],
		"VARIANT":                variant,
		"HOST_ID":                m.config.HostID,
		// TODO: Add pushgateway URL support
		// "METRICS_PUSHGATEWAY_URL": m.config.PushgatewayURL,
		"BATCH_TIMEOUT":          "1s",
		"BATCH_SIZE":             "1000",
	}

	// Merge with provided variables
	for k, v := range vars {
		allVars[k] = v
	}

	// Create template
	tmpl, err := template.New("config").Parse(config)
	if err != nil {
		return "", fmt.Errorf("failed to parse config template: %w", err)
	}

	// Execute template
	var buf strings.Builder
	if err := tmpl.Execute(&buf, allVars); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func (m *CollectorManager) monitorProcess(process *Process, logFile *os.File) {
	defer logFile.Close()

	// Wait for process to exit
	err := process.Cmd.Wait()

	m.mu.Lock()
	delete(m.processes, process.ID)
	m.mu.Unlock()

	if err != nil {
		log.Error().
			Err(err).
			Str("id", process.ID).
			Int("pid", process.Pid).
			Msg("Collector process exited with error")
	} else {
		log.Info().
			Str("id", process.ID).
			Int("pid", process.Pid).
			Msg("Collector process exited normally")
	}
}