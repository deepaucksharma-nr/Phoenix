package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	generatorv1 "github.com/phoenix-vnext/platform/api/proto/v1"
	"github.com/phoenix-vnext/platform/projects/generator/internal/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockConfigManager is a mock implementation of the ConfigManager interface
type MockConfigManager struct {
	mock.Mock
}

func (m *MockConfigManager) GenerateConfig(ctx context.Context, req config.GenerateRequest) (*config.GeneratedConfig, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*config.GeneratedConfig), args.Error(1)
}

func (m *MockConfigManager) ValidateConfig(ctx context.Context, cfg string) error {
	args := m.Called(ctx, cfg)
	return args.Error(0)
}

func (m *MockConfigManager) GetTemplate(ctx context.Context, name string) (*config.Template, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*config.Template), args.Error(1)
}

func (m *MockConfigManager) ListTemplates(ctx context.Context) ([]*config.Template, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*config.Template), args.Error(1)
}

func (m *MockConfigManager) CreateTemplate(ctx context.Context, tmpl *config.Template) error {
	args := m.Called(ctx, tmpl)
	return args.Error(0)
}

func (m *MockConfigManager) UpdateTemplate(ctx context.Context, name string, tmpl *config.Template) error {
	args := m.Called(ctx, name, tmpl)
	return args.Error(0)
}

func (m *MockConfigManager) DeleteTemplate(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func TestGeneratorServer_GenerateConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		request     *generatorv1.GenerateConfigurationRequest
		mockSetup   func(*MockConfigManager)
		wantErr     bool
		wantErrCode codes.Code
	}{
		{
			name: "successful generation",
			request: &generatorv1.GenerateConfigurationRequest{
				ExperimentId: "exp-123",
				Template:     "baseline",
				Parameters: map[string]string{
					"sampling_rate": "0.1",
					"batch_size":    "1000",
				},
			},
			mockSetup: func(m *MockConfigManager) {
				m.On("GenerateConfig", mock.Anything, mock.MatchedBy(func(req config.GenerateRequest) bool {
					return req.ExperimentID == "exp-123" &&
						req.Template == "baseline" &&
						req.Parameters["sampling_rate"] == "0.1" &&
						req.Parameters["batch_size"] == "1000"
				})).Return(&config.GeneratedConfig{
					ID:           "config-123",
					ExperimentID: "exp-123",
					Content:      "generated config content",
					Version:      "v1.0.0",
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "missing experiment ID",
			request: &generatorv1.GenerateConfigurationRequest{
				Template: "baseline",
			},
			mockSetup: func(m *MockConfigManager) {},
			wantErr:   true,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "missing template",
			request: &generatorv1.GenerateConfigurationRequest{
				ExperimentId: "exp-123",
			},
			mockSetup: func(m *MockConfigManager) {},
			wantErr:   true,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "generation error",
			request: &generatorv1.GenerateConfigurationRequest{
				ExperimentId: "exp-123",
				Template:     "baseline",
			},
			mockSetup: func(m *MockConfigManager) {
				m.On("GenerateConfig", mock.Anything, mock.Anything).
					Return(nil, assert.AnError)
			},
			wantErr:     true,
			wantErrCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockManager := new(MockConfigManager)
			tt.mockSetup(mockManager)

			server := NewGeneratorServer(mockManager)
			resp, err := server.GenerateConfiguration(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantErrCode, st.Code())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "config-123", resp.ConfigId)
				assert.Equal(t, "generated config content", resp.Configuration)
				assert.Equal(t, "v1.0.0", resp.Version)
			}

			mockManager.AssertExpectations(t)
		})
	}
}

func TestGeneratorServer_ValidateConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		request     *generatorv1.ValidateConfigurationRequest
		mockSetup   func(*MockConfigManager)
		wantErr     bool
		wantErrCode codes.Code
		wantValid   bool
	}{
		{
			name: "valid configuration",
			request: &generatorv1.ValidateConfigurationRequest{
				Configuration: "valid config content",
			},
			mockSetup: func(m *MockConfigManager) {
				m.On("ValidateConfig", mock.Anything, "valid config content").
					Return(nil)
			},
			wantErr:   false,
			wantValid: true,
		},
		{
			name: "invalid configuration",
			request: &generatorv1.ValidateConfigurationRequest{
				Configuration: "invalid config",
			},
			mockSetup: func(m *MockConfigManager) {
				m.On("ValidateConfig", mock.Anything, "invalid config").
					Return(assert.AnError)
			},
			wantErr:   false,
			wantValid: false,
		},
		{
			name: "empty configuration",
			request: &generatorv1.ValidateConfigurationRequest{
				Configuration: "",
			},
			mockSetup: func(m *MockConfigManager) {},
			wantErr:   true,
			wantErrCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockManager := new(MockConfigManager)
			tt.mockSetup(mockManager)

			server := NewGeneratorServer(mockManager)
			resp, err := server.ValidateConfiguration(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantErrCode, st.Code())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.wantValid, resp.Valid)
				if !tt.wantValid {
					assert.NotEmpty(t, resp.Errors)
				}
			}

			mockManager.AssertExpectations(t)
		})
	}
}

func TestGeneratorServer_GetTemplate(t *testing.T) {
	tests := []struct {
		name        string
		request     *generatorv1.GetTemplateRequest
		mockSetup   func(*MockConfigManager)
		wantErr     bool
		wantErrCode codes.Code
	}{
		{
			name: "successful get",
			request: &generatorv1.GetTemplateRequest{
				Name: "baseline",
			},
			mockSetup: func(m *MockConfigManager) {
				m.On("GetTemplate", mock.Anything, "baseline").
					Return(&config.Template{
						Name:        "baseline",
						Description: "Baseline template",
						Content:     "template content",
						Version:     "v1.0.0",
					}, nil)
			},
			wantErr: false,
		},
		{
			name: "template not found",
			request: &generatorv1.GetTemplateRequest{
				Name: "nonexistent",
			},
			mockSetup: func(m *MockConfigManager) {
				m.On("GetTemplate", mock.Anything, "nonexistent").
					Return(nil, assert.AnError)
			},
			wantErr:     true,
			wantErrCode: codes.NotFound,
		},
		{
			name: "empty name",
			request: &generatorv1.GetTemplateRequest{
				Name: "",
			},
			mockSetup: func(m *MockConfigManager) {},
			wantErr:   true,
			wantErrCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockManager := new(MockConfigManager)
			tt.mockSetup(mockManager)

			server := NewGeneratorServer(mockManager)
			resp, err := server.GetTemplate(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantErrCode, st.Code())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Template)
				assert.Equal(t, "baseline", resp.Template.Name)
				assert.Equal(t, "Baseline template", resp.Template.Description)
			}

			mockManager.AssertExpectations(t)
		})
	}
}

func TestGeneratorServer_ListTemplates(t *testing.T) {
	mockManager := new(MockConfigManager)
	mockManager.On("ListTemplates", mock.Anything).Return([]*config.Template{
		{
			Name:        "baseline",
			Description: "Baseline template",
			Version:     "v1.0.0",
		},
		{
			Name:        "optimized",
			Description: "Optimized template",
			Version:     "v2.0.0",
		},
	}, nil)

	server := NewGeneratorServer(mockManager)
	resp, err := server.ListTemplates(context.Background(), &generatorv1.ListTemplatesRequest{})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Templates, 2)
	assert.Equal(t, "baseline", resp.Templates[0].Name)
	assert.Equal(t, "optimized", resp.Templates[1].Name)

	mockManager.AssertExpectations(t)
}

func TestGeneratorServer_CreateTemplate(t *testing.T) {
	tests := []struct {
		name        string
		request     *generatorv1.CreateTemplateRequest
		mockSetup   func(*MockConfigManager)
		wantErr     bool
		wantErrCode codes.Code
	}{
		{
			name: "successful creation",
			request: &generatorv1.CreateTemplateRequest{
				Template: &generatorv1.Template{
					Name:        "new-template",
					Description: "New template",
					Content:     "template content",
				},
			},
			mockSetup: func(m *MockConfigManager) {
				m.On("CreateTemplate", mock.Anything, mock.MatchedBy(func(tmpl *config.Template) bool {
					return tmpl.Name == "new-template" &&
						tmpl.Description == "New template" &&
						tmpl.Content == "template content"
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "missing template",
			request: &generatorv1.CreateTemplateRequest{
				Template: nil,
			},
			mockSetup: func(m *MockConfigManager) {},
			wantErr:   true,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "empty template name",
			request: &generatorv1.CreateTemplateRequest{
				Template: &generatorv1.Template{
					Name:        "",
					Description: "Template",
					Content:     "content",
				},
			},
			mockSetup: func(m *MockConfigManager) {},
			wantErr:   true,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "creation error",
			request: &generatorv1.CreateTemplateRequest{
				Template: &generatorv1.Template{
					Name:        "new-template",
					Description: "New template",
					Content:     "content",
				},
			},
			mockSetup: func(m *MockConfigManager) {
				m.On("CreateTemplate", mock.Anything, mock.Anything).
					Return(assert.AnError)
			},
			wantErr:     true,
			wantErrCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockManager := new(MockConfigManager)
			tt.mockSetup(mockManager)

			server := NewGeneratorServer(mockManager)
			resp, err := server.CreateTemplate(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantErrCode, st.Code())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}

			mockManager.AssertExpectations(t)
		})
	}
}