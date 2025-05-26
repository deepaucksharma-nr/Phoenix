package grpc

import (
	"context"

	generatorv1 "github.com/phoenix-vnext/platform/api/proto/v1"
	"github.com/phoenix-vnext/platform/cmd/generator/internal/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GeneratorServer implements the gRPC generator service
type GeneratorServer struct {
	generatorv1.UnimplementedGeneratorServiceServer
	configManager config.ConfigManager
}

// NewGeneratorServer creates a new generator server
func NewGeneratorServer(configManager config.ConfigManager) *GeneratorServer {
	return &GeneratorServer{
		configManager: configManager,
	}
}

// GenerateConfiguration generates a configuration based on request parameters
func (s *GeneratorServer) GenerateConfiguration(ctx context.Context, req *generatorv1.GenerateConfigurationRequest) (*generatorv1.GenerateConfigurationResponse, error) {
	if req.ExperimentId == "" {
		return nil, status.Error(codes.InvalidArgument, "experiment_id is required")
	}
	if req.Template == "" {
		return nil, status.Error(codes.InvalidArgument, "template is required")
	}

	configReq := config.GenerateRequest{
		ExperimentID: req.ExperimentId,
		Template:     req.Template,
		Parameters:   req.Parameters,
	}

	generated, err := s.configManager.GenerateConfig(ctx, configReq)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate configuration")
	}

	return &generatorv1.GenerateConfigurationResponse{
		ConfigId:      generated.ID,
		Configuration: generated.Content,
		Version:       generated.Version,
	}, nil
}

// ValidateConfiguration validates a given configuration
func (s *GeneratorServer) ValidateConfiguration(ctx context.Context, req *generatorv1.ValidateConfigurationRequest) (*generatorv1.ValidateConfigurationResponse, error) {
	if req.Configuration == "" {
		return nil, status.Error(codes.InvalidArgument, "configuration is required")
	}

	err := s.configManager.ValidateConfig(ctx, req.Configuration)
	valid := err == nil
	var errors []string
	if !valid {
		errors = []string{err.Error()}
	}

	return &generatorv1.ValidateConfigurationResponse{
		Valid:  valid,
		Errors: errors,
	}, nil
}

// GetTemplate retrieves a specific template by name
func (s *GeneratorServer) GetTemplate(ctx context.Context, req *generatorv1.GetTemplateRequest) (*generatorv1.GetTemplateResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "template name is required")
	}

	template, err := s.configManager.GetTemplate(ctx, req.Name)
	if err != nil {
		return nil, status.Error(codes.NotFound, "template not found")
	}

	return &generatorv1.GetTemplateResponse{
		Template: &generatorv1.Template{
			Name:        template.Name,
			Description: template.Description,
			Content:     template.Content,
			Version:     template.Version,
		},
	}, nil
}

// ListTemplates returns all available templates
func (s *GeneratorServer) ListTemplates(ctx context.Context, req *generatorv1.ListTemplatesRequest) (*generatorv1.ListTemplatesResponse, error) {
	templates, err := s.configManager.ListTemplates(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list templates")
	}

	var protoTemplates []*generatorv1.Template
	for _, template := range templates {
		protoTemplates = append(protoTemplates, &generatorv1.Template{
			Name:        template.Name,
			Description: template.Description,
			Content:     template.Content,
			Version:     template.Version,
		})
	}

	return &generatorv1.ListTemplatesResponse{
		Templates: protoTemplates,
	}, nil
}

// CreateTemplate creates a new template
func (s *GeneratorServer) CreateTemplate(ctx context.Context, req *generatorv1.CreateTemplateRequest) (*generatorv1.CreateTemplateResponse, error) {
	if req.Template == nil {
		return nil, status.Error(codes.InvalidArgument, "template is required")
	}
	if req.Template.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "template name is required")
	}

	template := &config.Template{
		Name:        req.Template.Name,
		Description: req.Template.Description,
		Content:     req.Template.Content,
		Version:     req.Template.Version,
	}

	err := s.configManager.CreateTemplate(ctx, template)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create template")
	}

	return &generatorv1.CreateTemplateResponse{}, nil
}