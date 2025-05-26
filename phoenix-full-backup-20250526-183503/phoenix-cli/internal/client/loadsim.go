package client

import (
	"context"
	"time"
)

// LoadSimulationClient provides operations for load simulations
type LoadSimulationClient struct {
	apiClient *APIClient
}

// NewLoadSimulationClient creates a new load simulation client
func NewLoadSimulationClient(apiClient *APIClient) *LoadSimulationClient {
	return &LoadSimulationClient{
		apiClient: apiClient,
	}
}

// LoadSimulation represents a load simulation
type LoadSimulation struct {
	Name         string                           `json:"name"`
	ExperimentID string                           `json:"experiment_id"`
	Profile      string                           `json:"profile"`
	Duration     string                           `json:"duration"`
	ProcessCount int32                            `json:"process_count"`
	Status       string                           `json:"status"`
	StartTime    *time.Time                       `json:"start_time,omitempty"`
	EndTime      *time.Time                       `json:"end_time,omitempty"`
	Message      string                           `json:"message,omitempty"`
}

// CreateLoadSimulationRequest represents a request to create a load simulation
type CreateLoadSimulationRequest struct {
	ExperimentID string            `json:"experiment_id"`
	Profile      string            `json:"profile"`
	Duration     string            `json:"duration"`
	ProcessCount int32             `json:"process_count"`
	NodeSelector map[string]string `json:"node_selector,omitempty"`
}

// Start creates and starts a new load simulation
func (c *LoadSimulationClient) Start(ctx context.Context, req CreateLoadSimulationRequest) (*LoadSimulation, error) {
	resp, err := c.apiClient.doRequest("POST", "/api/v1/loadsimulations", req)
	if err != nil {
		return nil, err
	}

	var result LoadSimulation
	if err := c.apiClient.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Stop stops a running load simulation
func (c *LoadSimulationClient) Stop(ctx context.Context, name string) error {
	resp, err := c.apiClient.doRequest("DELETE", "/api/v1/loadsimulations/"+name, nil)
	if err != nil {
		return err
	}
	return c.apiClient.parseResponse(resp, nil)
}

// Get retrieves the status of a load simulation
func (c *LoadSimulationClient) Get(ctx context.Context, name string) (*LoadSimulation, error) {
	resp, err := c.apiClient.doRequest("GET", "/api/v1/loadsimulations/"+name, nil)
	if err != nil {
		return nil, err
	}

	var result LoadSimulation
	if err := c.apiClient.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// List lists all load simulations, optionally filtered by experiment ID
func (c *LoadSimulationClient) List(ctx context.Context, experimentID string) ([]LoadSimulation, error) {
	path := "/api/v1/loadsimulations"
	if experimentID != "" {
		path += "?experiment_id=" + experimentID
	}

	resp, err := c.apiClient.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		LoadSimulations []LoadSimulation `json:"load_simulations"`
	}
	if err := c.apiClient.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.LoadSimulations, nil
}

// GetProfiles returns available load simulation profiles
func (c *LoadSimulationClient) GetProfiles(ctx context.Context) ([]LoadProfile, error) {
	resp, err := c.apiClient.doRequest("GET", "/api/v1/loadsimulations/profiles", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Profiles []LoadProfile `json:"profiles"`
	}
	if err := c.apiClient.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Profiles, nil
}

// LoadProfile represents a load simulation profile
type LoadProfile struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	ProcessCount int32  `json:"default_process_count"`
	ChurnRate    float64 `json:"churn_rate"`
	CPUPattern   string `json:"cpu_pattern"`
	MemPattern   string `json:"mem_pattern"`
}