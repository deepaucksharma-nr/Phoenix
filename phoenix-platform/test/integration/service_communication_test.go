// +build integration

package integration

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/phoenix/platform/cmd/controller/internal/controller"
	controllergrpc "github.com/phoenix/platform/cmd/controller/internal/grpc"
	"github.com/phoenix/platform/cmd/generator/internal/config"
	generatorgrpc "github.com/phoenix/platform/cmd/generator/internal/grpc"
	controlgrpc "github.com/phoenix/platform/cmd/control-service/internal/grpc"
	pb "github.com/phoenix/platform/pkg/api/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

const bufSize = 1024 * 1024

// TestServiceCommunication tests gRPC communication between services
func TestServiceCommunication(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set RUN_INTEGRATION_TESTS=true to run.")
	}

	logger := zap.NewNop()

	// Test Experiment Controller
	t.Run("ExperimentController", func(t *testing.T) {
		conn, closer := setupExperimentService(t, logger)
		defer closer()

		client := pb.NewExperimentServiceClient(conn)
		testExperimentOperations(t, client)
	})

	// Test Config Generator
	t.Run("ConfigGenerator", func(t *testing.T) {
		conn, closer := setupGeneratorService(t, logger)
		defer closer()

		client := pb.NewGeneratorServiceClient(conn)
		testGeneratorOperations(t, client)
	})

	// Test Control Service
	t.Run("ControlService", func(t *testing.T) {
		conn, closer := setupControlService(t, logger)
		defer closer()

		client := pb.NewControllerServiceClient(conn)
		testControlOperations(t, client)
	})
}

func setupExperimentService(t *testing.T, logger *zap.Logger) (*grpc.ClientConn, func()) {
	// Create in-memory store
	store := NewMockExperimentStore()
	
	// Create controller
	expController := controller.NewExperimentController(logger, store)
	
	// Create gRPC server
	server := controllergrpc.NewSimpleExperimentServer(logger, expController)
	
	// Setup in-memory gRPC connection
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterExperimentServiceServer(s, server)
	
	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Error("Server exited with error", zap.Error(err))
		}
	}()

	// Create client connection
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", 
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	return conn, func() {
		conn.Close()
		s.Stop()
		lis.Close()
	}
}

func setupGeneratorService(t *testing.T, logger *zap.Logger) (*grpc.ClientConn, func()) {
	// Create mock config manager
	manager := NewMockConfigManager()
	
	// Create gRPC server
	server := generatorgrpc.NewGeneratorServer(manager)
	
	// Setup in-memory gRPC connection
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterGeneratorServiceServer(s, server)
	
	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Error("Server exited with error", zap.Error(err))
		}
	}()

	// Create client connection
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", 
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	return conn, func() {
		conn.Close()
		s.Stop()
		lis.Close()
	}
}

func setupControlService(t *testing.T, logger *zap.Logger) (*grpc.ClientConn, func()) {
	// Create gRPC server
	server := controlgrpc.NewControllerServer()
	
	// Setup in-memory gRPC connection
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterControllerServiceServer(s, server)
	
	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Error("Server exited with error", zap.Error(err))
		}
	}()

	// Create client connection
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", 
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	return conn, func() {
		conn.Close()
		s.Stop()
		lis.Close()
	}
}

func testExperimentOperations(t *testing.T, client pb.ExperimentServiceClient) {
	ctx := context.Background()

	// Create experiment
	createReq := &pb.CreateExperimentRequest{
		Name:              "Test Experiment",
		Description:       "Integration test experiment",
		BaselinePipeline:  "baseline-v1",
		CandidatePipeline: "candidate-v1",
		TargetNodes: map[string]string{
			"node1": "active",
			"node2": "active",
		},
	}

	createResp, err := client.CreateExperiment(ctx, createReq)
	require.NoError(t, err)
	assert.NotNil(t, createResp.Experiment)
	assert.NotEmpty(t, createResp.Experiment.Id)
	assert.Equal(t, "Test Experiment", createResp.Experiment.Name)

	expID := createResp.Experiment.Id

	// Get experiment
	getReq := &pb.GetExperimentRequest{Id: expID}
	getResp, err := client.GetExperiment(ctx, getReq)
	require.NoError(t, err)
	assert.Equal(t, expID, getResp.Experiment.Id)

	// List experiments
	listReq := &pb.ListExperimentsRequest{}
	listResp, err := client.ListExperiments(ctx, listReq)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(listResp.Experiments), 1)

	// Get status
	statusReq := &pb.GetExperimentStatusRequest{Id: expID}
	statusResp, err := client.GetExperimentStatus(ctx, statusReq)
	require.NoError(t, err)
	assert.NotEmpty(t, statusResp.Status)
}

func testGeneratorOperations(t *testing.T, client pb.GeneratorServiceClient) {
	ctx := context.Background()

	// Create template
	createReq := &pb.CreateTemplateRequest{
		Template: &pb.Template{
			Name:        "test-template",
			Description: "Test template",
			Content:     "collectors:\n  otlp:\n    protocols:\n      grpc:",
		},
	}

	createResp, err := client.CreateTemplate(ctx, createReq)
	require.NoError(t, err)
	assert.NotNil(t, createResp.Template)

	// Generate configuration
	genReq := &pb.GenerateConfigurationRequest{
		ExperimentId: "exp-123",
		Template:     "test-template",
		Parameters: map[string]string{
			"sampling_rate": "0.1",
		},
	}

	genResp, err := client.GenerateConfiguration(ctx, genReq)
	require.NoError(t, err)
	assert.NotEmpty(t, genResp.ConfigId)
	assert.NotEmpty(t, genResp.Configuration)

	// Validate configuration
	valReq := &pb.ValidateConfigurationRequest{
		Configuration: genResp.Configuration,
	}

	valResp, err := client.ValidateConfiguration(ctx, valReq)
	require.NoError(t, err)
	assert.True(t, valResp.Valid)
}

func testControlOperations(t *testing.T, client pb.ControllerServiceClient) {
	ctx := context.Background()

	// Apply control signal
	applyReq := &pb.ApplyControlSignalRequest{
		ExperimentId: "exp-123",
		Signal: &pb.ControlSignal{
			Type: pb.SignalType_SIGNAL_TYPE_TRAFFIC_SPLIT,
			Parameters: map[string]*structpb.Value{
				"baseline_weight":  structpb.NewNumberValue(70),
				"candidate_weight": structpb.NewNumberValue(30),
			},
		},
	}

	applyResp, err := client.ApplyControlSignal(ctx, applyReq)
	require.NoError(t, err)
	assert.NotEmpty(t, applyResp.SignalId)
	assert.Equal(t, pb.ControlStatus_CONTROL_STATUS_ACTIVE, applyResp.Status)

	// Get drift report
	driftReq := &pb.GetDriftReportRequest{
		ExperimentId: "exp-123",
	}

	driftResp, err := client.GetDriftReport(ctx, driftReq)
	require.NoError(t, err)
	assert.NotNil(t, driftResp.Report)
	assert.Equal(t, "exp-123", driftResp.Report.ExperimentId)
}

// Mock implementations for testing

type MockExperimentStore struct {
	experiments map[string]*controller.Experiment
}

func NewMockExperimentStore() *MockExperimentStore {
	return &MockExperimentStore{
		experiments: make(map[string]*controller.Experiment),
	}
}

func (m *MockExperimentStore) CreateExperiment(ctx context.Context, exp *controller.Experiment) error {
	m.experiments[exp.ID] = exp
	return nil
}

func (m *MockExperimentStore) GetExperiment(ctx context.Context, id string) (*controller.Experiment, error) {
	exp, ok := m.experiments[id]
	if !ok {
		return nil, fmt.Errorf("experiment not found")
	}
	return exp, nil
}

func (m *MockExperimentStore) UpdateExperiment(ctx context.Context, exp *controller.Experiment) error {
	m.experiments[exp.ID] = exp
	return nil
}

func (m *MockExperimentStore) ListExperiments(ctx context.Context, filter controller.ExperimentFilter) ([]*controller.Experiment, error) {
	var exps []*controller.Experiment
	for _, exp := range m.experiments {
		exps = append(exps, exp)
	}
	return exps, nil
}

type MockConfigManager struct {
	templates map[string]*config.Template
}

func NewMockConfigManager() *MockConfigManager {
	return &MockConfigManager{
		templates: make(map[string]*config.Template),
	}
}

func (m *MockConfigManager) GenerateConfig(ctx context.Context, req config.GenerateRequest) (*config.GeneratedConfig, error) {
	return &config.GeneratedConfig{
		ID:           "config-" + time.Now().Format("20060102150405"),
		ExperimentID: req.ExperimentID,
		Content:      "generated config content",
		Version:      "v1.0.0",
	}, nil
}

func (m *MockConfigManager) ValidateConfig(ctx context.Context, cfg string) error {
	if cfg == "" {
		return fmt.Errorf("empty configuration")
	}
	return nil
}

func (m *MockConfigManager) GetTemplate(ctx context.Context, name string) (*config.Template, error) {
	tmpl, ok := m.templates[name]
	if !ok {
		return nil, fmt.Errorf("template not found")
	}
	return tmpl, nil
}

func (m *MockConfigManager) ListTemplates(ctx context.Context) ([]*config.Template, error) {
	var tmpls []*config.Template
	for _, tmpl := range m.templates {
		tmpls = append(tmpls, tmpl)
	}
	return tmpls, nil
}

func (m *MockConfigManager) CreateTemplate(ctx context.Context, tmpl *config.Template) error {
	m.templates[tmpl.Name] = tmpl
	return nil
}

func (m *MockConfigManager) UpdateTemplate(ctx context.Context, name string, tmpl *config.Template) error {
	m.templates[name] = tmpl
	return nil
}

func (m *MockConfigManager) DeleteTemplate(ctx context.Context, name string) error {
	delete(m.templates, name)
	return nil
}