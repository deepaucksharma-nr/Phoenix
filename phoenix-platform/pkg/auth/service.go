package auth

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Service handles authentication
type Service struct {
	jwtSecret string
}

// NewService creates a new auth service
func NewService(jwtSecret string) *Service {
	return &Service{
		jwtSecret: jwtSecret,
	}
}

// UnaryInterceptor provides gRPC unary interceptor for authentication
func UnaryInterceptor(authService *Service) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// TODO: Implement JWT validation
		return handler(ctx, req)
	}
}

// StreamInterceptor provides gRPC stream interceptor for authentication
func StreamInterceptor(authService *Service) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// TODO: Implement JWT validation
		return handler(srv, ss)
	}
}

// ValidateToken validates a JWT token
func (s *Service) ValidateToken(ctx context.Context, token string) error {
	if token == "" {
		return errors.New("token required")
	}
	// TODO: Implement JWT validation
	return nil
}

// ExtractToken extracts token from context
func ExtractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("no metadata in context")
	}
	
	tokens := md.Get("authorization")
	if len(tokens) == 0 {
		return "", errors.New("no authorization header")
	}
	
	return tokens[0], nil
}