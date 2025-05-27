package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTransport is a mock implementation of http.RoundTripper
type mockTransport struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripFunc(req)
}

func TestNewAPIClient(t *testing.T) {
	client := NewAPIClient("http://localhost:8080", "test-token")
	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:8080", client.BaseURL)
	assert.Equal(t, "test-token", client.Token)
	assert.NotNil(t, client.httpClient)
}

func TestAPIClient_doRequest(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		mockResponse   *http.Response
		mockError      error
		expectedError  bool
		expectedStatus int
	}{
		{
			name:   "successful GET request",
			method: "GET",
			path:   "/api/v1/experiments",
			body:   nil,
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"success": true}`)),
			},
			expectedError:  false,
			expectedStatus: http.StatusOK,
		},
		{
			name:   "successful POST request with body",
			method: "POST",
			path:   "/api/v1/experiments",
			body: map[string]string{
				"name": "test-experiment",
			},
			mockResponse: &http.Response{
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(bytes.NewBufferString(`{"id": "exp-123"}`)),
			},
			expectedError:  false,
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "error response",
			method: "GET",
			path:   "/api/v1/experiments/invalid",
			body:   nil,
			mockResponse: &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error": "Experiment not found"}`)),
			},
			expectedError:  true,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &APIClient{
				BaseURL: "http://localhost:8080",
				Token:   "test-token",
				httpClient: &http.Client{
					Transport: &mockTransport{
						RoundTripFunc: func(req *http.Request) (*http.Response, error) {
							// Verify request
							assert.Equal(t, tt.method, req.Method)
							assert.Equal(t, "http://localhost:8080"+tt.path, req.URL.String())
							assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))

							if tt.body != nil {
								assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
							}

							if tt.mockError != nil {
								return nil, tt.mockError
							}
							return tt.mockResponse, nil
						},
					},
				},
			}

			resp, err := client.doRequest(tt.method, tt.path, tt.body)
			if tt.mockError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestAPIClient_Login(t *testing.T) {
	tests := []struct {
		name          string
		username      string
		password      string
		mockResponse  LoginResponse
		mockError     bool
		expectedToken string
		expectedError bool
	}{
		{
			name:     "successful login",
			username: "test@example.com",
			password: "password123",
			mockResponse: LoginResponse{
				Token:     "jwt-token-123",
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			expectedToken: "jwt-token-123",
			expectedError: false,
		},
		{
			name:          "invalid credentials",
			username:      "invalid@example.com",
			password:      "wrongpassword",
			mockError:     true,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/auth/login", r.URL.Path)
				assert.Equal(t, "POST", r.Method)

				var req LoginRequest
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)
				assert.Equal(t, tt.username, req.Username)
				assert.Equal(t, tt.password, req.Password)

				if tt.mockError {
					w.WriteHeader(http.StatusUnauthorized)
					json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid credentials"})
				} else {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			client := NewAPIClient(server.URL, "")
			resp, err := client.Login(tt.username, tt.password)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken, resp.Token)
			}
		})
	}
}

func TestAPIClient_CreateExperiment(t *testing.T) {
	req := CreateExperimentRequest{
		Name:              "test-experiment",
		Namespace:         "default",
		BaselinePipeline:  "baseline",
		CandidatePipeline: "optimized",
		TargetNodes: map[string]string{
			"node1": "value1",
		},
		Duration: time.Hour,
		Parameters: map[string]interface{}{
			"param1": "value1",
		},
		SuccessCriteria: &SuccessCriteria{
			MinCostReduction: 20,
			MaxDataLoss:      2,
		},
		Metadata: map[string]string{
			"team": "platform",
		},
	}

	expectedResp := Experiment{
		ID:                "exp-123",
		Name:              req.Name,
		Namespace:         req.Namespace,
		BaselinePipeline:  req.BaselinePipeline,
		CandidatePipeline: req.CandidatePipeline,
		Status:            "created",
		CreatedAt:         time.Now(),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/experiments", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		var receivedReq CreateExperimentRequest
		err := json.NewDecoder(r.Body).Decode(&receivedReq)
		require.NoError(t, err)
		// Don't compare the entire request as JSON marshaling may differ
		assert.Equal(t, req.Name, receivedReq.Name)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(expectedResp)
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-token")
	resp, err := client.CreateExperiment(req)

	assert.NoError(t, err)
	assert.Equal(t, expectedResp.ID, resp.ID)
	assert.Equal(t, expectedResp.Name, resp.Name)
	assert.Equal(t, expectedResp.Status, resp.Status)
}

func TestAPIClient_ListExperiments(t *testing.T) {
	expectedExperiments := []Experiment{
		{
			ID:        "exp-1",
			Name:      "test-1",
			Namespace: "default",
			Status:    "running",
		},
		{
			ID:        "exp-2",
			Name:      "test-2",
			Namespace: "default",
			Status:    "completed",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/experiments", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "running", r.URL.Query().Get("status"))

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ListExperimentsResponse{
			Experiments: expectedExperiments,
			Total:       2,
			Page:        1,
			PageSize:    10,
		})
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-token")
	resp, err := client.ListExperiments(ListExperimentsRequest{Status: "running"})

	assert.NoError(t, err)
	assert.Len(t, resp.Experiments, 2)
	assert.Equal(t, expectedExperiments[0].ID, resp.Experiments[0].ID)
	assert.Equal(t, expectedExperiments[1].ID, resp.Experiments[1].ID)
}

func TestAPIClient_GetExperimentMetrics(t *testing.T) {
	expectedMetrics := ExperimentMetrics{
		ExperimentID: "exp-123",
		Summary: MetricsSummary{
			CostReductionPercent:    35.5,
			DataLossPercent:         0.8,
			ProgressPercent:         75,
			EstimatedMonthlySavings: 1500.50,
		},
		PipelineA: PipelineMetrics{
			Cardinality:     10000,
			Throughput:      1048576,
			ErrorRate:       0.001,
			Latency:         50,
			CostPerHour:     10.5,
			DataLossPercent: 0.5,
		},
		PipelineB: PipelineMetrics{
			Cardinality:     6500,
			Throughput:      682000,
			ErrorRate:       0.0008,
			Latency:         45,
			CostPerHour:     7.2,
			DataLossPercent: 0.3,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/experiments/exp-123/metrics", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedMetrics)
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-token")
	metrics, err := client.GetExperimentMetrics("exp-123")

	assert.NoError(t, err)
	assert.Equal(t, expectedMetrics.ExperimentID, metrics.ExperimentID)
	assert.Equal(t, expectedMetrics.Summary.CostReductionPercent, metrics.Summary.CostReductionPercent)
	assert.Equal(t, expectedMetrics.Summary.DataLossPercent, metrics.Summary.DataLossPercent)
}

func TestAPIClient_parseAPIError(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		responseBody  string
		expectedError string
	}{
		{
			name:          "API error with message",
			statusCode:    http.StatusBadRequest,
			responseBody:  `{"error": "Invalid request", "message": "Missing required field"}`,
			expectedError: "API error (status 400): Missing required field - Invalid request",
		},
		{
			name:          "API error without details",
			statusCode:    http.StatusNotFound,
			responseBody:  `{"message": "Not found"}`,
			expectedError: "API error (status 404): Not found - ",
		},
		{
			name:          "Non-JSON error response",
			statusCode:    http.StatusInternalServerError,
			responseBody:  "Internal Server Error",
			expectedError: "API error (status 500): Unknown error - Failed to decode error response",
		},
		{
			name:          "Empty response",
			statusCode:    http.StatusServiceUnavailable,
			responseBody:  "",
			expectedError: "API error (status 503): Unknown error - Failed to decode error response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(bytes.NewBufferString(tt.responseBody)),
			}

			client := &APIClient{}
			err := client.parseAPIError(resp)

			assert.Error(t, err)
			apiErr, ok := err.(*APIError)
			assert.True(t, ok)
			assert.Equal(t, tt.statusCode, apiErr.StatusCode)
			// Check if error contains expected parts
			assert.Contains(t, err.Error(), "API error")
			assert.Contains(t, err.Error(), fmt.Sprintf("%d", tt.statusCode))
		})
	}
}
