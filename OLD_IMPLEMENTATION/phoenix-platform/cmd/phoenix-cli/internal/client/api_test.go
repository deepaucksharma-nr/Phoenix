package client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockHTTPClient is a mock implementation of HTTPClient
type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
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
				httpClient: &mockHTTPClient{
					DoFunc: func(req *http.Request) (*http.Response, error) {
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
			}

			resp, err := client.doRequest(tt.method, tt.path, tt.body)
			if tt.expectedError {
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
					json.NewEncoder(w).Encode(APIError{Message: "Invalid credentials"})
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
		Name:          "test-experiment",
		Namespace:     "default",
		PipelineA:     "baseline",
		PipelineB:     "optimized",
		TrafficSplit:  "50/50",
		Duration:      "1h",
		Selector:      "app=test",
		SuccessCriteria: SuccessCriteria{
			MinCostReduction: 20,
			MaxDataLoss:      2,
		},
		Metadata: map[string]string{
			"team": "platform",
		},
	}

	expectedResp := Experiment{
		ID:        "exp-123",
		Name:      req.Name,
		Namespace: req.Namespace,
		Status:    "created",
		CreatedAt: time.Now(),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/experiments", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		var receivedReq CreateExperimentRequest
		err := json.NewDecoder(r.Body).Decode(&receivedReq)
		require.NoError(t, err)
		assert.Equal(t, req, receivedReq)

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
		assert.Equal(t, "default", r.URL.Query().Get("namespace"))
		assert.Equal(t, "running", r.URL.Query().Get("status"))

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ListExperimentsResponse{
			Experiments: expectedExperiments,
		})
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-token")
	resp, err := client.ListExperiments("default", "running")

	assert.NoError(t, err)
	assert.Len(t, resp.Experiments, 2)
	assert.Equal(t, expectedExperiments[0].ID, resp.Experiments[0].ID)
	assert.Equal(t, expectedExperiments[1].ID, resp.Experiments[1].ID)
}

func TestAPIClient_GetExperimentMetrics(t *testing.T) {
	expectedMetrics := ExperimentMetrics{
		ExperimentID: "exp-123",
		Summary: MetricsSummary{
			CostReductionPercent:   35.5,
			DataLossPercent:        0.8,
			ProgressPercent:        75,
			EstimatedMonthlySavings: 1500.50,
		},
		PipelineA: PipelineMetrics{
			DataPointsPerSecond: 10000,
			BytesPerSecond:      1048576,
			ErrorRate:           0.001,
		},
		PipelineB: PipelineMetrics{
			DataPointsPerSecond: 6500,
			BytesPerSecond:      682000,
			ErrorRate:           0.0008,
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
			responseBody:  `{"error": "Invalid request", "details": "Missing required field"}`,
			expectedError: "API error (400): Invalid request - Missing required field",
		},
		{
			name:          "API error without details",
			statusCode:    http.StatusNotFound,
			responseBody:  `{"error": "Not found"}`,
			expectedError: "API error (404): Not found",
		},
		{
			name:          "Non-JSON error response",
			statusCode:    http.StatusInternalServerError,
			responseBody:  "Internal Server Error",
			expectedError: "API error (500): Internal Server Error",
		},
		{
			name:          "Empty response",
			statusCode:    http.StatusServiceUnavailable,
			responseBody:  "",
			expectedError: "API error (503): Service Unavailable",
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
			assert.Equal(t, tt.expectedError, err.Error())
		})
	}
}