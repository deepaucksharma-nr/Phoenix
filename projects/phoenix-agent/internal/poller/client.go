package poller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/phoenix/platform/projects/phoenix-agent/internal/config"
	"github.com/rs/zerolog/log"
)

type Client struct {
	config     *config.Config
	httpClient *http.Client
}

type Task struct {
	ID           string                 `json:"id"`
	HostID       string                 `json:"host_id"`
	ExperimentID string                 `json:"experiment_id"`
	Type         string                 `json:"type"`
	Action       string                 `json:"action"`
	Config       map[string]interface{} `json:"config"`
	Priority     int                    `json:"priority"`
}

type AgentStatus struct {
	HostID        string                 `json:"host_id"`
	AgentVersion  string                 `json:"agent_version"`
	Status        string                 `json:"status"`
	ActiveTasks   []string               `json:"active_tasks"`
	ResourceUsage ResourceUsage          `json:"resource_usage"`
}

type ResourceUsage struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float64 `json:"memory_percent"`
	MemoryBytes   int64   `json:"memory_bytes"`
	DiskPercent   float64 `json:"disk_percent"`
	DiskBytes     int64   `json:"disk_bytes"`
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 35 * time.Second, // Slightly longer than server's 30s long-poll
		},
	}
}

// GetTasks polls the API for pending tasks
func (c *Client) GetTasks(ctx context.Context) ([]*Task, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.config.GetAPIEndpoint("/tasks"), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Agent-Host-ID", c.config.HostID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var tasks []*Task
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return nil, fmt.Errorf("failed to decode tasks: %w", err)
	}

	return tasks, nil
}

// UpdateTaskStatus updates the status of a task
func (c *Client) UpdateTaskStatus(ctx context.Context, taskID, status string, result map[string]interface{}, errorMessage string) error {
	payload := map[string]interface{}{
		"status": status,
	}
	if result != nil {
		payload["result"] = result
	}
	if errorMessage != "" {
		payload["error_message"] = errorMessage
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/%s/status", c.config.GetAPIEndpoint("/tasks"), taskID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Agent-Host-ID", c.config.HostID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// SendHeartbeat sends agent status to the API
func (c *Client) SendHeartbeat(ctx context.Context, status *AgentStatus) error {
	status.HostID = c.config.HostID
	status.AgentVersion = "1.0.0" // TODO: Make this configurable

	data, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.GetAPIEndpoint("/heartbeat"), bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Agent-Host-ID", c.config.HostID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send heartbeat: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// SendMetrics sends collected metrics to the API
func (c *Client) SendMetrics(ctx context.Context, metrics []map[string]interface{}) error {
	payload := map[string]interface{}{
		"timestamp": time.Now(),
		"metrics":   metrics,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.GetAPIEndpoint("/metrics"), bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Agent-Host-ID", c.config.HostID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// SendLogs sends logs to the API
func (c *Client) SendLogs(ctx context.Context, taskID string, logs []LogEntry) error {
	payload := map[string]interface{}{
		"task_id": taskID,
		"logs":    logs,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal logs: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.GetAPIEndpoint("/logs"), bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Agent-Host-ID", c.config.HostID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send logs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}