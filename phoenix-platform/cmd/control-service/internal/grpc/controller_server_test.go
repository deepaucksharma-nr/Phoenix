package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	controllerv1 "github.com/phoenix/platform/api/phoenix/controller/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestControllerServer_GetControlSignal(t *testing.T) {
	server := NewControllerServer()

	// Test getting non-existent signal
	resp, err := server.GetControlSignal(context.Background(), &controllerv1.GetControlSignalRequest{
		SignalId: "non-existent",
	})
	
	assert.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.Nil(t, resp)
}

func TestControllerServer_ApplyControlSignal(t *testing.T) {
	tests := []struct {
		name        string
		request     *controllerv1.ApplyControlSignalRequest
		wantErr     bool
		wantErrCode codes.Code
		checkResp   func(*testing.T, *controllerv1.ApplyControlSignalResponse)
	}{
		{
			name: "successful traffic split",
			request: &controllerv1.ApplyControlSignalRequest{
				ExperimentId: "exp-123",
				Signal: &controllerv1.ControlSignal{
					Type: controllerv1.SignalType_SIGNAL_TYPE_TRAFFIC_SPLIT,
					Parameters: map[string]*structpb.Value{
						"baseline_weight":  structpb.NewNumberValue(50),
						"candidate_weight": structpb.NewNumberValue(50),
					},
				},
			},
			wantErr: false,
			checkResp: func(t *testing.T, resp *controllerv1.ApplyControlSignalResponse) {
				assert.NotEmpty(t, resp.SignalId)
				assert.Equal(t, controllerv1.ControlStatus_CONTROL_STATUS_ACTIVE, resp.Status)
			},
		},
		{
			name: "successful rollback",
			request: &controllerv1.ApplyControlSignalRequest{
				ExperimentId: "exp-123",
				Signal: &controllerv1.ControlSignal{
					Type: controllerv1.SignalType_SIGNAL_TYPE_ROLLBACK,
					Parameters: map[string]*structpb.Value{
						"reason": structpb.NewStringValue("high error rate"),
					},
				},
			},
			wantErr: false,
			checkResp: func(t *testing.T, resp *controllerv1.ApplyControlSignalResponse) {
				assert.NotEmpty(t, resp.SignalId)
				assert.Equal(t, controllerv1.ControlStatus_CONTROL_STATUS_ACTIVE, resp.Status)
			},
		},
		{
			name: "successful config update",
			request: &controllerv1.ApplyControlSignalRequest{
				ExperimentId: "exp-123",
				Signal: &controllerv1.ControlSignal{
					Type: controllerv1.SignalType_SIGNAL_TYPE_CONFIG_UPDATE,
					Parameters: map[string]*structpb.Value{
						"config_id": structpb.NewStringValue("config-456"),
						"version":   structpb.NewStringValue("v2.0.0"),
					},
				},
			},
			wantErr: false,
			checkResp: func(t *testing.T, resp *controllerv1.ApplyControlSignalResponse) {
				assert.NotEmpty(t, resp.SignalId)
				assert.Equal(t, controllerv1.ControlStatus_CONTROL_STATUS_ACTIVE, resp.Status)
			},
		},
		{
			name: "missing experiment ID",
			request: &controllerv1.ApplyControlSignalRequest{
				Signal: &controllerv1.ControlSignal{
					Type: controllerv1.SignalType_SIGNAL_TYPE_TRAFFIC_SPLIT,
				},
			},
			wantErr:     true,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "missing signal",
			request: &controllerv1.ApplyControlSignalRequest{
				ExperimentId: "exp-123",
			},
			wantErr:     true,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "unspecified signal type",
			request: &controllerv1.ApplyControlSignalRequest{
				ExperimentId: "exp-123",
				Signal: &controllerv1.ControlSignal{
					Type: controllerv1.SignalType_SIGNAL_TYPE_UNSPECIFIED,
				},
			},
			wantErr:     true,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "traffic split missing weights",
			request: &controllerv1.ApplyControlSignalRequest{
				ExperimentId: "exp-123",
				Signal: &controllerv1.ControlSignal{
					Type: controllerv1.SignalType_SIGNAL_TYPE_TRAFFIC_SPLIT,
					Parameters: map[string]*structpb.Value{},
				},
			},
			wantErr:     true,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "traffic split invalid weights",
			request: &controllerv1.ApplyControlSignalRequest{
				ExperimentId: "exp-123",
				Signal: &controllerv1.ControlSignal{
					Type: controllerv1.SignalType_SIGNAL_TYPE_TRAFFIC_SPLIT,
					Parameters: map[string]*structpb.Value{
						"baseline_weight":  structpb.NewNumberValue(60),
						"candidate_weight": structpb.NewNumberValue(50), // Sum > 100
					},
				},
			},
			wantErr:     true,
			wantErrCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewControllerServer()
			resp, err := server.ApplyControlSignal(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantErrCode, st.Code())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.checkResp != nil {
					tt.checkResp(t, resp)
				}
			}
		})
	}
}

func TestControllerServer_GetDriftReport(t *testing.T) {
	server := NewControllerServer()

	// Create a control signal to have some data
	_, err := server.ApplyControlSignal(context.Background(), &controllerv1.ApplyControlSignalRequest{
		ExperimentId: "exp-123",
		Signal: &controllerv1.ControlSignal{
			Type: controllerv1.SignalType_SIGNAL_TYPE_TRAFFIC_SPLIT,
			Parameters: map[string]*structpb.Value{
				"baseline_weight":  structpb.NewNumberValue(50),
				"candidate_weight": structpb.NewNumberValue(50),
			},
		},
	})
	assert.NoError(t, err)

	tests := []struct {
		name        string
		request     *controllerv1.GetDriftReportRequest
		wantErr     bool
		wantErrCode codes.Code
	}{
		{
			name: "successful drift report",
			request: &controllerv1.GetDriftReportRequest{
				ExperimentId: "exp-123",
			},
			wantErr: false,
		},
		{
			name: "missing experiment ID",
			request: &controllerv1.GetDriftReportRequest{
				ExperimentId: "",
			},
			wantErr:     true,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "experiment not found",
			request: &controllerv1.GetDriftReportRequest{
				ExperimentId: "non-existent",
			},
			wantErr:     true,
			wantErrCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.GetDriftReport(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantErrCode, st.Code())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Report)
				assert.Equal(t, "exp-123", resp.Report.ExperimentId)
				assert.NotNil(t, resp.Report.Timestamp)
				assert.GreaterOrEqual(t, resp.Report.DriftScore, 0.0)
				assert.LessOrEqual(t, resp.Report.DriftScore, 1.0)
			}
		})
	}
}

func TestControllerServer_ListControlSignals(t *testing.T) {
	server := NewControllerServer()

	// Create some control signals
	signals := []struct {
		expID string
		sType controllerv1.SignalType
	}{
		{"exp-123", controllerv1.SignalType_SIGNAL_TYPE_TRAFFIC_SPLIT},
		{"exp-123", controllerv1.SignalType_SIGNAL_TYPE_CONFIG_UPDATE},
		{"exp-456", controllerv1.SignalType_SIGNAL_TYPE_ROLLBACK},
	}

	for _, s := range signals {
		_, err := server.ApplyControlSignal(context.Background(), &controllerv1.ApplyControlSignalRequest{
			ExperimentId: s.expID,
			Signal: &controllerv1.ControlSignal{
				Type:       s.sType,
				Parameters: map[string]*structpb.Value{},
			},
		})
		assert.NoError(t, err)
	}

	tests := []struct {
		name         string
		request      *controllerv1.ListControlSignalsRequest
		wantErr      bool
		wantErrCode  codes.Code
		expectedCount int
	}{
		{
			name: "list all signals for experiment",
			request: &controllerv1.ListControlSignalsRequest{
				ExperimentId: "exp-123",
			},
			wantErr:       false,
			expectedCount: 2,
		},
		{
			name: "list signals for different experiment",
			request: &controllerv1.ListControlSignalsRequest{
				ExperimentId: "exp-456",
			},
			wantErr:       false,
			expectedCount: 1,
		},
		{
			name: "missing experiment ID",
			request: &controllerv1.ListControlSignalsRequest{
				ExperimentId: "",
			},
			wantErr:     true,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "no signals for experiment",
			request: &controllerv1.ListControlSignalsRequest{
				ExperimentId: "exp-789",
			},
			wantErr:       false,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.ListControlSignals(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantErrCode, st.Code())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Len(t, resp.Signals, tt.expectedCount)
			}
		})
	}
}

func TestControllerServer_ValidateControlSignal(t *testing.T) {
	server := NewControllerServer()

	tests := []struct {
		name    string
		signal  *controllerv1.ControlSignal
		wantErr bool
		errors  []string
	}{
		{
			name: "valid traffic split",
			signal: &controllerv1.ControlSignal{
				Type: controllerv1.SignalType_SIGNAL_TYPE_TRAFFIC_SPLIT,
				Parameters: map[string]*structpb.Value{
					"baseline_weight":  structpb.NewNumberValue(70),
					"candidate_weight": structpb.NewNumberValue(30),
				},
			},
			wantErr: false,
		},
		{
			name: "unspecified type",
			signal: &controllerv1.ControlSignal{
				Type: controllerv1.SignalType_SIGNAL_TYPE_UNSPECIFIED,
			},
			wantErr: true,
			errors:  []string{"signal type is required"},
		},
		{
			name: "traffic split weights not summing to 100",
			signal: &controllerv1.ControlSignal{
				Type: controllerv1.SignalType_SIGNAL_TYPE_TRAFFIC_SPLIT,
				Parameters: map[string]*structpb.Value{
					"baseline_weight":  structpb.NewNumberValue(60),
					"candidate_weight": structpb.NewNumberValue(60),
				},
			},
			wantErr: true,
			errors:  []string{"traffic weights must sum to 100"},
		},
		{
			name: "negative weight",
			signal: &controllerv1.ControlSignal{
				Type: controllerv1.SignalType_SIGNAL_TYPE_TRAFFIC_SPLIT,
				Parameters: map[string]*structpb.Value{
					"baseline_weight":  structpb.NewNumberValue(-10),
					"candidate_weight": structpb.NewNumberValue(110),
				},
			},
			wantErr: true,
			errors:  []string{"baseline_weight must be between 0 and 100"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := server.validateControlSignal(tt.signal)
			
			if tt.wantErr {
				assert.NotEmpty(t, errors)
				for i, expectedErr := range tt.errors {
					if i < len(errors) {
						assert.Contains(t, errors[i].Message, expectedErr)
					}
				}
			} else {
				assert.Empty(t, errors)
			}
		})
	}
}

func TestControllerServer_ConcurrentAccess(t *testing.T) {
	server := NewControllerServer()
	ctx := context.Background()

	// Test concurrent writes
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			_, err := server.ApplyControlSignal(ctx, &controllerv1.ApplyControlSignalRequest{
				ExperimentId: "exp-concurrent",
				Signal: &controllerv1.ControlSignal{
					Type: controllerv1.SignalType_SIGNAL_TYPE_CONFIG_UPDATE,
					Parameters: map[string]*structpb.Value{
						"config_id": structpb.NewStringValue("config-" + string(rune(id))),
					},
				},
			})
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all signals were created
	resp, err := server.ListControlSignals(ctx, &controllerv1.ListControlSignalsRequest{
		ExperimentId: "exp-concurrent",
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Signals, 10)
}