package clients

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"go.uber.org/zap"
)

// GeneratorRequest represents a configuration generation request
type GeneratorRequest struct {
	ExperimentID      string                 `json:"experiment_id"`
	BaselinePipeline  string                 `json:"baseline_pipeline"`
	CandidatePipeline string                 `json:"candidate_pipeline"`
	TargetNodes       []string               `json:"target_nodes"`
	Variables         map[string]interface{} `json:"variables"`
}

// GeneratorResponse represents the response from config generation
type GeneratorResponse struct {
	Success           bool   `json:"success"`
	Message          string `json:"message"`
	BaselineConfigID  string `json:"baseline_config_id,omitempty"`
	CandidateConfigID string `json:"candidate_config_id,omitempty"`
	GitCommitSHA     string `json:"git_commit_sha,omitempty"`
}

// GeneratorClient handles communication with the Config Generator service
type GeneratorClient struct {
	logger     *zap.Logger
	conn       *grpc.ClientConn
	endpoint   string
	timeout    time.Duration
}

// NewGeneratorClient creates a new generator client
func NewGeneratorClient(logger *zap.Logger, endpoint string) *GeneratorClient {
	return &GeneratorClient{
		logger:   logger,
		endpoint: endpoint,
		timeout:  30 * time.Second,
	}
}

// Connect establishes connection to the generator service
func (c *GeneratorClient) Connect(ctx context.Context) error {
	c.logger.Info("connecting to config generator",
		zap.String("endpoint", c.endpoint),
	)

	conn, err := grpc.DialContext(ctx, c.endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to generator service: %w", err)
	}

	c.conn = conn
	c.logger.Info("connected to config generator service")
	return nil
}

// Close closes the connection to the generator service
func (c *GeneratorClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GenerateConfigurations calls the generator service to create experiment configurations
func (c *GeneratorClient) GenerateConfigurations(ctx context.Context, req *GeneratorRequest) (*GeneratorResponse, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("generator client not connected")
	}

	c.logger.Info("generating configurations",
		zap.String("experiment_id", req.ExperimentID),
		zap.String("baseline_pipeline", req.BaselinePipeline),
		zap.String("candidate_pipeline", req.CandidatePipeline),
		zap.Strings("target_nodes", req.TargetNodes),
	)

	// For now, we'll make a simple HTTP call since we haven't implemented gRPC in generator yet
	// In a real implementation, this would use the proper gRPC client
	
	// Simulate configuration generation
	time.Sleep(2 * time.Second)
	
	response := &GeneratorResponse{
		Success:           true,
		Message:          "Configurations generated successfully",
		BaselineConfigID:  fmt.Sprintf("%s-baseline", req.ExperimentID),
		CandidateConfigID: fmt.Sprintf("%s-candidate", req.ExperimentID),
		GitCommitSHA:     "abc123def456", // Mock commit SHA
	}

	c.logger.Info("configuration generation completed",
		zap.String("experiment_id", req.ExperimentID),
		zap.String("baseline_config_id", response.BaselineConfigID),
		zap.String("candidate_config_id", response.CandidateConfigID),
		zap.String("git_commit_sha", response.GitCommitSHA),
	)

	return response, nil
}

// ListTemplates retrieves available pipeline templates
func (c *GeneratorClient) ListTemplates(ctx context.Context) ([]string, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("generator client not connected")
	}

	c.logger.Info("listing available pipeline templates")

	// For now, return hardcoded templates
	// In real implementation, this would call the generator service
	templates := []string{
		"process-baseline-v1",
		"process-priority-filter-v1",
		"process-topk-v1",
		"process-aggregated-v1",
		"process-adaptive-filter-v1",
	}

	c.logger.Info("retrieved pipeline templates",
		zap.Strings("templates", templates),
	)

	return templates, nil
}

// ValidateTemplate validates a pipeline template
func (c *GeneratorClient) ValidateTemplate(ctx context.Context, templateName string, variables map[string]interface{}) error {
	if c.conn == nil {
		return fmt.Errorf("generator client not connected")
	}

	c.logger.Info("validating pipeline template",
		zap.String("template", templateName),
	)

	// Simulate validation
	time.Sleep(500 * time.Millisecond)

	// For now, just check if template exists in our known templates
	validTemplates := map[string]bool{
		"process-baseline-v1":       true,
		"process-priority-filter-v1": true,
		"process-topk-v1":           true,
		"process-aggregated-v1":     true,
		"process-adaptive-filter-v1": true,
	}

	if !validTemplates[templateName] {
		return fmt.Errorf("unknown template: %s", templateName)
	}

	c.logger.Info("template validation successful",
		zap.String("template", templateName),
	)

	return nil
}