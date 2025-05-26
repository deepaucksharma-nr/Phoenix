package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// APIClient handles communication with the Phoenix API
type APIClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewAPIClient creates a new API client
func NewAPIClient(baseURL, token string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request with authentication
func (c *APIClient) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// parseResponse parses the HTTP response into the provided interface
func (c *APIClient) parseResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error != "" {
			return fmt.Errorf("API error: %s", errorResp.Error)
		}
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// CreateExperiment creates a new experiment
func (c *APIClient) CreateExperiment(req CreateExperimentRequest) (*Experiment, error) {
	resp, err := c.doRequest("POST", "/api/v1/experiments", req)
	if err != nil {
		return nil, err
	}

	var result struct {
		Experiment Experiment `json:"experiment"`
	}
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Experiment, nil
}

// ListExperiments lists experiments with optional filters
func (c *APIClient) ListExperiments(req ListExperimentsRequest) ([]Experiment, error) {
	// Build query parameters
	params := url.Values{}
	if req.Status != "" {
		params.Add("status", req.Status)
	}
	if req.PageSize > 0 {
		params.Add("page_size", fmt.Sprintf("%d", req.PageSize))
	}

	path := "/api/v1/experiments"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Experiments []Experiment `json:"experiments"`
	}
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Experiments, nil
}

// GetExperiment gets a single experiment by ID
func (c *APIClient) GetExperiment(id string) (*Experiment, error) {
	resp, err := c.doRequest("GET", "/api/v1/experiments/"+id, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Experiment Experiment `json:"experiment"`
	}
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Experiment, nil
}

// StartExperiment starts an experiment
func (c *APIClient) StartExperiment(id string) error {
	resp, err := c.doRequest("POST", "/api/v1/experiments/"+id+"/start", nil)
	if err != nil {
		return err
	}
	return c.parseResponse(resp, nil)
}

// StopExperiment stops an experiment
func (c *APIClient) StopExperiment(id string) error {
	resp, err := c.doRequest("POST", "/api/v1/experiments/"+id+"/stop", nil)
	if err != nil {
		return err
	}
	return c.parseResponse(resp, nil)
}

// PromoteExperiment promotes an experiment variant
func (c *APIClient) PromoteExperiment(id string, variant string) error {
	req := struct {
		Variant string `json:"variant"`
	}{
		Variant: variant,
	}

	resp, err := c.doRequest("POST", "/api/v1/experiments/"+id+"/promote", req)
	if err != nil {
		return err
	}
	return c.parseResponse(resp, nil)
}

// GetExperimentMetrics gets metrics for an experiment
func (c *APIClient) GetExperimentMetrics(id string) (*ExperimentMetrics, error) {
	resp, err := c.doRequest("GET", "/api/v1/experiments/"+id+"/metrics", nil)
	if err != nil {
		return nil, err
	}

	var metrics ExperimentMetrics
	if err := c.parseResponse(resp, &metrics); err != nil {
		return nil, err
	}

	return &metrics, nil
}

// CheckExperimentOverlap checks for overlapping experiments
func (c *APIClient) CheckExperimentOverlap(req CreateExperimentRequest) (*OverlapResult, error) {
	checkReq := struct {
		CreateExperimentRequest
		CheckOnly bool `json:"check_only"`
	}{
		CreateExperimentRequest: req,
		CheckOnly:               true,
	}

	resp, err := c.doRequest("POST", "/api/v1/experiments/check-overlap", checkReq)
	if err != nil {
		return nil, err
	}

	var result OverlapResult
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ListPipelines lists available pipeline templates
func (c *APIClient) ListPipelines() ([]Pipeline, error) {
	resp, err := c.doRequest("GET", "/api/v1/pipelines", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Pipelines []Pipeline `json:"pipelines"`
	}
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Pipelines, nil
}

// CreatePipelineDeployment creates a new pipeline deployment
func (c *APIClient) CreatePipelineDeployment(req CreatePipelineDeploymentRequest) (*PipelineDeployment, error) {
	resp, err := c.doRequest("POST", "/api/v1/pipelines/deployments", req)
	if err != nil {
		return nil, err
	}

	var deployment PipelineDeployment
	if err := c.parseResponse(resp, &deployment); err != nil {
		return nil, err
	}

	return &deployment, nil
}

// ListPipelineDeployments lists pipeline deployments
func (c *APIClient) ListPipelineDeployments(req ListPipelineDeploymentsRequest) ([]PipelineDeployment, error) {
	params := url.Values{}
	if req.Namespace != "" {
		params.Add("namespace", req.Namespace)
	}
	if req.Status != "" {
		params.Add("status", req.Status)
	}

	path := "/api/v1/pipelines/deployments"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Deployments []PipelineDeployment `json:"deployments"`
	}
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Deployments, nil
}

// GetPipelineDeployment gets a single pipeline deployment
func (c *APIClient) GetPipelineDeployment(deploymentID string) (*PipelineDeployment, error) {
	resp, err := c.doRequest("GET", "/api/v1/pipelines/deployments/"+deploymentID, nil)
	if err != nil {
		return nil, err
	}

	var deployment PipelineDeployment
	if err := c.parseResponse(resp, &deployment); err != nil {
		return nil, err
	}

	return &deployment, nil
}

// GetPipelineDeploymentStatus retrieves aggregated deployment status
func (c *APIClient) GetPipelineDeploymentStatus(deploymentID string) (*DeploymentStatusResponse, error) {
	resp, err := c.doRequest("GET", "/api/v1/pipelines/deployments/"+deploymentID+"/status", nil)
	if err != nil {
		return nil, err
	}

	var status DeploymentStatusResponse
	if err := c.parseResponse(resp, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

// GetPipelineConfig retrieves the active configuration from a deployment
func (c *APIClient) GetPipelineConfig(deploymentID string) (string, error) {
	resp, err := c.doRequest("GET", "/api/v1/pipelines/deployments/"+deploymentID+"/config", nil)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error != "" {
			return "", fmt.Errorf("API error: %s", errorResp.Error)
		}
		return "", fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Return raw YAML content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
}

// RollbackPipeline rolls back a pipeline deployment to a previous version
func (c *APIClient) RollbackPipeline(deploymentID string, req RollbackPipelineRequest) (*PipelineDeployment, error) {
	resp, err := c.doRequest("POST", "/api/v1/pipelines/deployments/"+deploymentID+"/rollback", req)
	if err != nil {
		return nil, err
	}

	var deployment PipelineDeployment
	if err := c.parseResponse(resp, &deployment); err != nil {
		return nil, err
	}

	return &deployment, nil
}

// DeletePipelineDeployment deletes a pipeline deployment
func (c *APIClient) DeletePipelineDeployment(deploymentID string) error {
	resp, err := c.doRequest("DELETE", "/api/v1/pipelines/deployments/"+deploymentID, nil)
	if err != nil {
		return err
	}
	return c.parseResponse(resp, nil)
}

// DeleteExperiment deletes an experiment
func (c *APIClient) DeleteExperiment(id string) error {
	resp, err := c.doRequest("DELETE", "/api/v1/experiments/"+id, nil)
	if err != nil {
		return err
	}
	return c.parseResponse(resp, nil)
}
