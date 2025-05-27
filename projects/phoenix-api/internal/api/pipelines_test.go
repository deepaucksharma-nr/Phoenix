package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/phoenix/platform/projects/phoenix-api/internal/services"
	"github.com/phoenix/platform/projects/phoenix-api/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestServer creates a test server with mocked dependencies
func setupTestServer(t *testing.T) *Server {
	// Create a mock store
	mockStore := &store.CompositeStore{}
	
	// Create services
	templateRenderer := services.NewPipelineTemplateRenderer()
	
	// Create server
	server := &Server{
		store:            mockStore,
		templateRenderer: templateRenderer,
		router:           chi.NewRouter(),
	}
	
	// Setup routes
	server.setupRoutes()
	
	return server
}

func TestHandleValidatePipeline(t *testing.T) {
	server := setupTestServer(t)

	tests := []struct {
		name           string
		request        interface{}
		expectedStatus int
		expectedValid  bool
		expectedError  string
	}{
		{
			name: "ValidCompleteConfig",
			request: map[string]interface{}{
				"config": map[string]interface{}{
					"receivers": map[string]interface{}{
						"otlp": map[string]interface{}{
							"protocols": map[string]interface{}{
								"grpc": map[string]interface{}{
									"endpoint": "0.0.0.0:4317",
								},
							},
						},
					},
					"processors": map[string]interface{}{
						"batch": map[string]interface{}{
							"timeout":         "1s",
							"send_batch_size": 1024,
						},
					},
					"exporters": map[string]interface{}{
						"prometheus": map[string]interface{}{
							"endpoint": "0.0.0.0:8889",
						},
					},
					"service": map[string]interface{}{
						"pipelines": map[string]interface{}{
							"metrics": map[string]interface{}{
								"receivers":  []string{"otlp"},
								"processors": []string{"batch"},
								"exporters":  []string{"prometheus"},
							},
						},
					},
				},
			},
			expectedStatus: http.StatusOK,
			expectedValid:  true,
		},
		{
			name: "ValidYAMLConfig",
			request: map[string]interface{}{
				"yaml": `
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
processors:
  batch:
    timeout: 1s
exporters:
  prometheus:
    endpoint: 0.0.0.0:8889
service:
  pipelines:
    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [prometheus]
`,
			},
			expectedStatus: http.StatusOK,
			expectedValid:  true,
		},
		{
			name: "MissingReceivers",
			request: map[string]interface{}{
				"config": map[string]interface{}{
					"exporters": map[string]interface{}{
						"prometheus": map[string]interface{}{
							"endpoint": "0.0.0.0:8889",
						},
					},
					"service": map[string]interface{}{
						"pipelines": map[string]interface{}{
							"metrics": map[string]interface{}{
								"exporters": []string{"prometheus"},
							},
						},
					},
				},
			},
			expectedStatus: http.StatusOK,
			expectedValid:  false,
			expectedError:  "pipeline must have at least one receiver",
		},
		{
			name: "MissingExporters",
			request: map[string]interface{}{
				"config": map[string]interface{}{
					"receivers": map[string]interface{}{
						"otlp": map[string]interface{}{
							"protocols": map[string]interface{}{
								"grpc": map[string]interface{}{
									"endpoint": "0.0.0.0:4317",
								},
							},
						},
					},
					"service": map[string]interface{}{
						"pipelines": map[string]interface{}{
							"metrics": map[string]interface{}{
								"receivers": []string{"otlp"},
							},
						},
					},
				},
			},
			expectedStatus: http.StatusOK,
			expectedValid:  false,
			expectedError:  "pipeline must have at least one exporter",
		},
		{
			name: "InvalidYAML",
			request: map[string]interface{}{
				"yaml": `
receivers:
  otlp:
    invalid yaml content [[[
`,
			},
			expectedStatus: http.StatusOK,
			expectedValid:  false,
			expectedError:  "Invalid YAML",
		},
		{
			name: "UndefinedProcessorReference",
			request: map[string]interface{}{
				"config": map[string]interface{}{
					"receivers": map[string]interface{}{
						"otlp": map[string]interface{}{
							"protocols": map[string]interface{}{
								"grpc": map[string]interface{}{
									"endpoint": "0.0.0.0:4317",
								},
							},
						},
					},
					"exporters": map[string]interface{}{
						"prometheus": map[string]interface{}{
							"endpoint": "0.0.0.0:8889",
						},
					},
					"service": map[string]interface{}{
						"pipelines": map[string]interface{}{
							"metrics": map[string]interface{}{
								"receivers":  []string{"otlp"},
								"processors": []string{"undefined_processor"},
								"exporters":  []string{"prometheus"},
							},
						},
					},
				},
			},
			expectedStatus: http.StatusOK,
			expectedValid:  false,
			expectedError:  "references undefined processor",
		},
		{
			name: "InvalidProcessorConfig",
			request: map[string]interface{}{
				"config": map[string]interface{}{
					"receivers": map[string]interface{}{
						"otlp": map[string]interface{}{
							"protocols": map[string]interface{}{
								"grpc": map[string]interface{}{
									"endpoint": "0.0.0.0:4317",
								},
							},
						},
					},
					"processors": map[string]interface{}{
						"batch": map[string]interface{}{
							"timeout":         "invalid-duration",
							"send_batch_size": -100,
						},
					},
					"exporters": map[string]interface{}{
						"prometheus": map[string]interface{}{
							"endpoint": "0.0.0.0:8889",
						},
					},
					"service": map[string]interface{}{
						"pipelines": map[string]interface{}{
							"metrics": map[string]interface{}{
								"receivers":  []string{"otlp"},
								"processors": []string{"batch"},
								"exporters":  []string{"prometheus"},
							},
						},
					},
				},
			},
			expectedStatus: http.StatusOK,
			expectedValid:  false,
			expectedError:  "invalid timeout",
		},
		{
			name: "EmptyOTLPEndpoint",
			request: map[string]interface{}{
				"config": map[string]interface{}{
					"receivers": map[string]interface{}{
						"otlp": map[string]interface{}{
							"protocols": map[string]interface{}{
								"grpc": map[string]interface{}{
									"endpoint": "",
								},
							},
						},
					},
					"exporters": map[string]interface{}{
						"prometheus": map[string]interface{}{
							"endpoint": "0.0.0.0:8889",
						},
					},
					"service": map[string]interface{}{
						"pipelines": map[string]interface{}{
							"metrics": map[string]interface{}{
								"receivers": []string{"otlp"},
								"exporters": []string{"prometheus"},
							},
						},
					},
				},
			},
			expectedStatus: http.StatusOK,
			expectedValid:  false,
			expectedError:  "gRPC endpoint cannot be empty",
		},
		{
			name: "PhoenixAdaptiveFilterValidation",
			request: map[string]interface{}{
				"config": map[string]interface{}{
					"receivers": map[string]interface{}{
						"otlp": map[string]interface{}{
							"protocols": map[string]interface{}{
								"grpc": map[string]interface{}{
									"endpoint": "0.0.0.0:4317",
								},
							},
						},
					},
					"processors": map[string]interface{}{
						"phoenix_adaptive_filter": map[string]interface{}{
							"adaptive_filter": map[string]interface{}{
								"enabled": true,
								"thresholds": map[string]interface{}{
									"cardinality_limit": -100,
								},
							},
						},
					},
					"exporters": map[string]interface{}{
						"prometheus": map[string]interface{}{
							"endpoint": "0.0.0.0:8889",
						},
					},
					"service": map[string]interface{}{
						"pipelines": map[string]interface{}{
							"metrics": map[string]interface{}{
								"receivers":  []string{"otlp"},
								"processors": []string{"phoenix_adaptive_filter"},
								"exporters":  []string{"prometheus"},
							},
						},
					},
				},
			},
			expectedStatus: http.StatusOK,
			expectedValid:  false,
			expectedError:  "cardinality_limit must be positive",
		},
		{
			name:           "NoConfigProvided",
			request:        map[string]interface{}{},
			expectedStatus: http.StatusOK,
			expectedValid:  false,
			expectedError:  "No configuration provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/api/v1/pipelines/validate", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			server.router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var response map[string]interface{}
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedValid, response["valid"])

			if tt.expectedError != "" {
				errorMsg, ok := response["error"].(string)
				assert.True(t, ok, "Expected error message in response")
				assert.Contains(t, errorMsg, tt.expectedError)
			}

			if tt.expectedValid {
				assert.Equal(t, "Pipeline configuration is valid", response["message"])
			}
		})
	}
}

func TestHandleRenderPipeline(t *testing.T) {
	server := setupTestServer(t)

	tests := []struct {
		name           string
		request        map[string]interface{}
		expectedStatus int
		shouldContain  []string
	}{
		{
			name: "RenderBaselineTemplate",
			request: map[string]interface{}{
				"template":      "baseline",
				"experiment_id": "exp-123",
				"variant":       "baseline",
				"host_id":       "host-456",
				"parameters": map[string]interface{}{
					"prometheus_endpoint": "http://prometheus:9090",
				},
			},
			expectedStatus: http.StatusOK,
			shouldContain: []string{
				"receivers:",
				"processors:",
				"exporters:",
				"experiment_id: \"exp-123\"",
				"variant: \"baseline\"",
				"host_id: \"host-456\"",
			},
		},
		{
			name: "MissingTemplate",
			request: map[string]interface{}{
				"experiment_id": "exp-123",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "InvalidTemplate",
			request: map[string]interface{}{
				"template": "non-existent-template",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/api/v1/pipelines/render", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			server.router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)

				rendered, ok := response["rendered"].(string)
				assert.True(t, ok, "Expected rendered config in response")

				for _, expected := range tt.shouldContain {
					assert.Contains(t, rendered, expected)
				}
			}
		})
	}
}