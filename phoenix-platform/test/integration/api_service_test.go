package integration

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/phoenix/platform/pkg/api"
	pb "github.com/phoenix/platform/pkg/api/v1"
	"github.com/phoenix/platform/pkg/store"
)

type APIServiceTestSuite struct {
	suite.Suite
	db         *sql.DB
	store      store.Store
	grpcServer *grpc.Server
	grpcConn   *grpc.ClientConn
	client     pb.ExperimentServiceClient
	port       int
	logger     *zap.Logger
}

func (suite *APIServiceTestSuite) SetupSuite() {
	// Initialize logger
	logger, _ := zap.NewDevelopment()
	suite.logger = logger

	// Setup test database
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://phoenix:phoenix@localhost/phoenix_test?sslmode=disable"
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	require.NoError(suite.T(), err)
	suite.db = db

	// Create store
	store, err := store.NewPostgresStore(dbURL)
	require.NoError(suite.T(), err)
	suite.store = store

	// Clear any existing data
	suite.cleanDatabase()

	// Create gRPC server
	suite.grpcServer = grpc.NewServer()
	
	// Register services
	experimentService := api.NewExperimentService(suite.store, nil, suite.logger)
	pb.RegisterExperimentServiceServer(suite.grpcServer, experimentService)

	// Start gRPC server on random port
	listener, err := net.Listen("tcp", ":0")
	require.NoError(suite.T(), err)
	suite.port = listener.Addr().(*net.TCPAddr).Port

	go func() {
		if err := suite.grpcServer.Serve(listener); err != nil {
			suite.logger.Error("gRPC server error", zap.Error(err))
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Create gRPC client
	conn, err := grpc.Dial(
		fmt.Sprintf("localhost:%d", suite.port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(suite.T(), err)
	suite.grpcConn = conn
	suite.client = pb.NewExperimentServiceClient(conn)
}

func (suite *APIServiceTestSuite) TearDownSuite() {
	// Stop gRPC server
	suite.grpcServer.GracefulStop()

	// Close connections
	if suite.grpcConn != nil {
		suite.grpcConn.Close()
	}
	if suite.store != nil {
		suite.store.Close()
	}
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *APIServiceTestSuite) SetupTest() {
	// Clean database before each test
	suite.cleanDatabase()
}

func (suite *APIServiceTestSuite) cleanDatabase() {
	queries := []string{
		"TRUNCATE TABLE experiments CASCADE",
		"TRUNCATE TABLE pipelines CASCADE",
		"TRUNCATE TABLE users CASCADE",
		"TRUNCATE TABLE audit_logs CASCADE",
	}

	for _, query := range queries {
		_, err := suite.db.Exec(query)
		if err != nil {
			// Table might not exist yet
			suite.logger.Debug("failed to truncate table", zap.Error(err))
		}
	}
}

// Test Cases

func (suite *APIServiceTestSuite) TestCreateExperiment() {
	ctx := context.Background()

	req := &pb.CreateExperimentRequest{
		Name:              "Test Experiment",
		Description:       "Integration test experiment",
		BaselinePipeline:  "baseline-v1",
		CandidatePipeline: "candidate-v1",
		TargetNodes:       []string{"node-1", "node-2"},
		Duration:          "1h",
		LoadProfile:       "realistic",
	}

	resp, err := suite.client.CreateExperiment(ctx, req)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.NotEmpty(suite.T(), resp.Experiment.Id)
	assert.Equal(suite.T(), req.Name, resp.Experiment.Name)
	assert.Equal(suite.T(), pb.ExperimentStatus_PENDING, resp.Experiment.Status)
}

func (suite *APIServiceTestSuite) TestCreateExperimentValidation() {
	ctx := context.Background()

	testCases := []struct {
		name    string
		request *pb.CreateExperimentRequest
		errMsg  string
	}{
		{
			name: "missing name",
			request: &pb.CreateExperimentRequest{
				BaselinePipeline:  "baseline",
				CandidatePipeline: "candidate",
				TargetNodes:       []string{"node-1"},
			},
			errMsg: "experiment name is required",
		},
		{
			name: "missing baseline pipeline",
			request: &pb.CreateExperimentRequest{
				Name:              "Test",
				CandidatePipeline: "candidate",
				TargetNodes:       []string{"node-1"},
			},
			errMsg: "baseline pipeline is required",
		},
		{
			name: "missing target nodes",
			request: &pb.CreateExperimentRequest{
				Name:              "Test",
				BaselinePipeline:  "baseline",
				CandidatePipeline: "candidate",
				TargetNodes:       []string{},
			},
			errMsg: "at least one target node is required",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, err := suite.client.CreateExperiment(ctx, tc.request)
			assert.Error(suite.T(), err)
			assert.Contains(suite.T(), err.Error(), tc.errMsg)
		})
	}
}

func (suite *APIServiceTestSuite) TestGetExperiment() {
	ctx := context.Background()

	// Create an experiment first
	createReq := &pb.CreateExperimentRequest{
		Name:              "Get Test Experiment",
		Description:       "Test get operation",
		BaselinePipeline:  "baseline-v1",
		CandidatePipeline: "candidate-v1",
		TargetNodes:       []string{"node-1"},
		Duration:          "30m",
	}

	createResp, err := suite.client.CreateExperiment(ctx, createReq)
	require.NoError(suite.T(), err)

	// Get the experiment
	getReq := &pb.GetExperimentRequest{
		Id: createResp.Experiment.Id,
	}

	getResp, err := suite.client.GetExperiment(ctx, getReq)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), getResp)
	assert.Equal(suite.T(), createResp.Experiment.Id, getResp.Experiment.Id)
	assert.Equal(suite.T(), createReq.Name, getResp.Experiment.Name)
}

func (suite *APIServiceTestSuite) TestGetNonExistentExperiment() {
	ctx := context.Background()

	req := &pb.GetExperimentRequest{
		Id: uuid.New().String(),
	}

	_, err := suite.client.GetExperiment(ctx, req)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "not found")
}

func (suite *APIServiceTestSuite) TestListExperiments() {
	ctx := context.Background()

	// Create multiple experiments
	for i := 0; i < 5; i++ {
		req := &pb.CreateExperimentRequest{
			Name:              fmt.Sprintf("List Test Experiment %d", i),
			BaselinePipeline:  "baseline-v1",
			CandidatePipeline: "candidate-v1",
			TargetNodes:       []string{"node-1"},
		}
		_, err := suite.client.CreateExperiment(ctx, req)
		require.NoError(suite.T(), err)
	}

	// List experiments
	listReq := &pb.ListExperimentsRequest{
		Limit: 3,
	}

	listResp, err := suite.client.ListExperiments(ctx, listReq)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), listResp.Experiments, 3)
	assert.Equal(suite.T(), int32(5), listResp.Total)
}

func (suite *APIServiceTestSuite) TestListExperimentsPagination() {
	ctx := context.Background()

	// Create experiments
	for i := 0; i < 10; i++ {
		req := &pb.CreateExperimentRequest{
			Name:              fmt.Sprintf("Page Test Experiment %02d", i),
			BaselinePipeline:  "baseline-v1",
			CandidatePipeline: "candidate-v1",
			TargetNodes:       []string{"node-1"},
		}
		_, err := suite.client.CreateExperiment(ctx, req)
		require.NoError(suite.T(), err)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Get first page
	page1Req := &pb.ListExperimentsRequest{
		Limit:  5,
		Offset: 0,
	}
	page1Resp, err := suite.client.ListExperiments(ctx, page1Req)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), page1Resp.Experiments, 5)

	// Get second page
	page2Req := &pb.ListExperimentsRequest{
		Limit:  5,
		Offset: 5,
	}
	page2Resp, err := suite.client.ListExperiments(ctx, page2Req)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), page2Resp.Experiments, 5)

	// Ensure no overlap
	page1IDs := make(map[string]bool)
	for _, exp := range page1Resp.Experiments {
		page1IDs[exp.Id] = true
	}
	for _, exp := range page2Resp.Experiments {
		assert.False(suite.T(), page1IDs[exp.Id], "Found duplicate experiment in pages")
	}
}

func (suite *APIServiceTestSuite) TestUpdateExperimentStatus() {
	ctx := context.Background()

	// Create experiment
	createReq := &pb.CreateExperimentRequest{
		Name:              "Status Update Test",
		BaselinePipeline:  "baseline-v1",
		CandidatePipeline: "candidate-v1",
		TargetNodes:       []string{"node-1"},
	}
	createResp, err := suite.client.CreateExperiment(ctx, createReq)
	require.NoError(suite.T(), err)

	// Update status
	updateReq := &pb.UpdateExperimentStatusRequest{
		Id:     createResp.Experiment.Id,
		Status: pb.ExperimentStatus_RUNNING,
	}
	updateResp, err := suite.client.UpdateExperimentStatus(ctx, updateReq)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), pb.ExperimentStatus_RUNNING, updateResp.Experiment.Status)
	assert.NotNil(suite.T(), updateResp.Experiment.StartedAt)
}

func (suite *APIServiceTestSuite) TestDeleteExperiment() {
	ctx := context.Background()

	// Create experiment
	createReq := &pb.CreateExperimentRequest{
		Name:              "Delete Test",
		BaselinePipeline:  "baseline-v1",
		CandidatePipeline: "candidate-v1",
		TargetNodes:       []string{"node-1"},
	}
	createResp, err := suite.client.CreateExperiment(ctx, createReq)
	require.NoError(suite.T(), err)

	// Delete experiment
	deleteReq := &pb.DeleteExperimentRequest{
		Id: createResp.Experiment.Id,
	}
	_, err = suite.client.DeleteExperiment(ctx, deleteReq)
	require.NoError(suite.T(), err)

	// Verify it's gone
	getReq := &pb.GetExperimentRequest{
		Id: createResp.Experiment.Id,
	}
	_, err = suite.client.GetExperiment(ctx, getReq)
	assert.Error(suite.T(), err)
}

func (suite *APIServiceTestSuite) TestConcurrentExperimentCreation() {
	ctx := context.Background()
	numGoroutines := 10
	errChan := make(chan error, numGoroutines)
	idChan := make(chan string, numGoroutines)

	// Create experiments concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			req := &pb.CreateExperimentRequest{
				Name:              fmt.Sprintf("Concurrent Test %d", index),
				BaselinePipeline:  "baseline-v1",
				CandidatePipeline: "candidate-v1",
				TargetNodes:       []string{"node-1"},
			}
			resp, err := suite.client.CreateExperiment(ctx, req)
			if err != nil {
				errChan <- err
			} else {
				idChan <- resp.Experiment.Id
			}
		}(i)
	}

	// Collect results
	ids := make(map[string]bool)
	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-errChan:
			suite.T().Fatalf("Concurrent creation failed: %v", err)
		case id := <-idChan:
			// Check for duplicate IDs
			assert.False(suite.T(), ids[id], "Duplicate ID generated")
			ids[id] = true
		case <-time.After(5 * time.Second):
			suite.T().Fatal("Timeout waiting for concurrent operations")
		}
	}

	assert.Len(suite.T(), ids, numGoroutines)
}

// Run the test suite
func TestAPIServiceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(APIServiceTestSuite))
}