package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/phoenix-vnext/platform/cmd/controller/internal/controller"
	pb "github.com/phoenix-vnext/platform/pkg/api/v1"
)

// MockExperimentStore is a mock implementation of the ExperimentStore interface
type MockExperimentStore struct {
	mock.Mock
}

func (m *MockExperimentStore) CreateExperiment(ctx context.Context, exp *controller.Experiment) error {
	args := m.Called(ctx, exp)
	return args.Error(0)
}

func (m *MockExperimentStore) GetExperiment(ctx context.Context, id string) (*controller.Experiment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*controller.Experiment), args.Error(1)
}

func (m *MockExperimentStore) UpdateExperiment(ctx context.Context, exp *controller.Experiment) error {
	args := m.Called(ctx, exp)
	return args.Error(0)
}

func (m *MockExperimentStore) ListExperiments(ctx context.Context, filter controller.ExperimentFilter) ([]*controller.Experiment, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*controller.Experiment), args.Error(1)
}

func TestCreateExperiment(t *testing.T) {
	tests := []struct {
		name          string
		request       *pb.CreateExperimentRequest
		mockSetup     func(*MockExperimentStore)
		expectedError bool
		errorMessage  string
	}{
		{
			name: "successful creation",
			request: &pb.CreateExperimentRequest{
				Name:              "Test Experiment",
				Description:       "Test Description",
				BaselinePipeline:  "baseline-v1",
				CandidatePipeline: "candidate-v1",
				TargetNodes: map[string]string{
					"node1": "active",
					"node2": "active",
				},
			},
			mockSetup: func(m *MockExperimentStore) {
				m.On("CreateExperiment", mock.Anything, mock.MatchedBy(func(exp *controller.Experiment) bool {
					return exp.Name == "Test Experiment" &&
						exp.Description == "Test Description" &&
						exp.Config.BaselinePipeline == "baseline-v1" &&
						exp.Config.CandidatePipeline == "candidate-v1" &&
						len(exp.Config.TargetHosts) == 2
				})).Return(nil)
				// Mock for background processExperiment goroutine
				m.On("GetExperiment", mock.Anything, mock.AnythingOfType("string")).Return(&controller.Experiment{
					ID:    "test-id",
					Phase: controller.ExperimentPhasePending,
				}, nil).Maybe()
				m.On("UpdateExperiment", mock.Anything, mock.Anything).Return(nil).Maybe()
			},
			expectedError: false,
		},
		{
			name: "missing name",
			request: &pb.CreateExperimentRequest{
				BaselinePipeline:  "baseline-v1",
				CandidatePipeline: "candidate-v1",
			},
			mockSetup:     func(m *MockExperimentStore) {},
			expectedError: true,
			errorMessage:  "experiment name is required",
		},
		{
			name: "missing baseline pipeline",
			request: &pb.CreateExperimentRequest{
				Name:              "Test Experiment",
				CandidatePipeline: "candidate-v1",
			},
			mockSetup:     func(m *MockExperimentStore) {},
			expectedError: true,
			errorMessage:  "baseline pipeline is required",
		},
		{
			name: "missing candidate pipeline",
			request: &pb.CreateExperimentRequest{
				Name:             "Test Experiment",
				BaselinePipeline: "baseline-v1",
			},
			mockSetup:     func(m *MockExperimentStore) {},
			expectedError: true,
			errorMessage:  "candidate pipeline is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := zap.NewNop()
			mockStore := new(MockExperimentStore)
			tt.mockSetup(mockStore)

			// Create controller with mock store
			expController := controller.NewExperimentController(logger, mockStore)
			server := NewSimpleExperimentServer(logger, expController)

			// Execute
			resp, err := server.CreateExperiment(context.Background(), tt.request)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Experiment)
				assert.NotEmpty(t, resp.Experiment.Id)
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestGetExperiment(t *testing.T) {
	tests := []struct {
		name          string
		request       *pb.GetExperimentRequest
		mockSetup     func(*MockExperimentStore)
		expectedError bool
		errorMessage  string
	}{
		{
			name: "successful retrieval",
			request: &pb.GetExperimentRequest{
				Id: "exp-123",
			},
			mockSetup: func(m *MockExperimentStore) {
				exp := &controller.Experiment{
					ID:          "exp-123",
					Name:        "Test Experiment",
					Description: "Test Description",
					Phase:       controller.ExperimentPhaseRunning,
				}
				m.On("GetExperiment", mock.Anything, "exp-123").Return(exp, nil)
			},
			expectedError: false,
		},
		{
			name: "missing id",
			request: &pb.GetExperimentRequest{
				Id: "",
			},
			mockSetup:     func(m *MockExperimentStore) {},
			expectedError: true,
			errorMessage:  "experiment ID is required",
		},
		{
			name: "experiment not found",
			request: &pb.GetExperimentRequest{
				Id: "exp-nonexistent",
			},
			mockSetup: func(m *MockExperimentStore) {
				m.On("GetExperiment", mock.Anything, "exp-nonexistent").Return(nil, assert.AnError)
			},
			expectedError: true,
			errorMessage:  "failed to retrieve experiment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := zap.NewNop()
			mockStore := new(MockExperimentStore)
			tt.mockSetup(mockStore)

			expController := controller.NewExperimentController(logger, mockStore)
			server := NewSimpleExperimentServer(logger, expController)

			// Execute
			resp, err := server.GetExperiment(context.Background(), tt.request)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Experiment)
				assert.Equal(t, "exp-123", resp.Experiment.Id)
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestListExperiments(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockStore := new(MockExperimentStore)

	experiments := []*controller.Experiment{
		{
			ID:          "exp-1",
			Name:        "Experiment 1",
			Description: "Description 1",
			Phase:       controller.ExperimentPhaseRunning,
		},
		{
			ID:          "exp-2",
			Name:        "Experiment 2",
			Description: "Description 2",
			Phase:       controller.ExperimentPhasePending,
		},
	}

	mockStore.On("ListExperiments", mock.Anything, mock.MatchedBy(func(filter controller.ExperimentFilter) bool {
		return filter.Limit == 50
	})).Return(experiments, nil)

	expController := controller.NewExperimentController(logger, mockStore)
	server := NewSimpleExperimentServer(logger, expController)

	// Execute
	resp, err := server.ListExperiments(context.Background(), &pb.ListExperimentsRequest{})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Experiments, 2)
	assert.Equal(t, "exp-1", resp.Experiments[0].Id)
	assert.Equal(t, "exp-2", resp.Experiments[1].Id)

	mockStore.AssertExpectations(t)
}

func TestGetExperimentStatus(t *testing.T) {
	tests := []struct {
		name          string
		request       *pb.GetExperimentStatusRequest
		mockSetup     func(*MockExperimentStore)
		expectedError bool
		errorMessage  string
	}{
		{
			name: "successful status retrieval",
			request: &pb.GetExperimentStatusRequest{
				Id: "exp-123",
			},
			mockSetup: func(m *MockExperimentStore) {
				exp := &controller.Experiment{
					ID:    "exp-123",
					Phase: controller.ExperimentPhaseRunning,
				}
				m.On("GetExperiment", mock.Anything, "exp-123").Return(exp, nil)
			},
			expectedError: false,
		},
		{
			name: "missing id",
			request: &pb.GetExperimentStatusRequest{
				Id: "",
			},
			mockSetup:     func(m *MockExperimentStore) {},
			expectedError: true,
			errorMessage:  "experiment ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := zap.NewNop()
			mockStore := new(MockExperimentStore)
			tt.mockSetup(mockStore)

			expController := controller.NewExperimentController(logger, mockStore)
			server := NewSimpleExperimentServer(logger, expController)

			// Execute
			resp, err := server.GetExperimentStatus(context.Background(), tt.request)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.Status)
			}

			mockStore.AssertExpectations(t)
		})
	}
}