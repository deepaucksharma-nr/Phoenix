package interfaces

import (
	"context"
	"time"
	
	"github.com/phoenix/platform/pkg/common/websocket"
)

// ServiceRegistry provides service discovery capabilities
// This interface enables services to find and communicate with each other
type ServiceRegistry interface {
	// Register registers a service instance
	Register(ctx context.Context, service *ServiceInstance) error
	
	// Deregister removes a service instance
	Deregister(ctx context.Context, serviceID string) error
	
	// Discover finds service instances by name
	Discover(ctx context.Context, serviceName string) ([]*ServiceInstance, error)
	
	// HealthCheck updates the health status of a service
	HealthCheck(ctx context.Context, serviceID string, status HealthStatus) error
	
	// Watch watches for changes to service instances
	Watch(ctx context.Context, serviceName string) (<-chan ServiceEvent, error)
}

// ServiceInstance represents a running service instance
type ServiceInstance struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Address     string            `json:"address"`
	Port        int               `json:"port"`
	Protocol    string            `json:"protocol"` // grpc, http, https
	HealthCheck *HealthCheckConfig `json:"health_check,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Status      HealthStatus      `json:"status"`
	RegisteredAt time.Time        `json:"registered_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

// HealthCheckConfig defines health check parameters
type HealthCheckConfig struct {
	Path     string        `json:"path"`
	Interval time.Duration `json:"interval"`
	Timeout  time.Duration `json:"timeout"`
	Retries  int           `json:"retries"`
}

// HealthStatus represents the health state of a service
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// ServiceEvent represents a change in service state
type ServiceEvent struct {
	Type     ServiceEventType `json:"type"`
	Service  *ServiceInstance `json:"service"`
	Timestamp time.Time       `json:"timestamp"`
}

// ServiceEventType defines types of service events
type ServiceEventType string

const (
	ServiceEventTypeRegistered   ServiceEventType = "registered"
	ServiceEventTypeDeregistered ServiceEventType = "deregistered"
	ServiceEventTypeHealthChange ServiceEventType = "health_changed"
)

// ServiceClient provides a generic interface for service communication
type ServiceClient interface {
	// Call makes a synchronous call to a service
	Call(ctx context.Context, service, method string, request, response interface{}) error
	
	// CallWithRetry makes a call with retry logic
	CallWithRetry(ctx context.Context, service, method string, request, response interface{}, retry *RetryConfig) error
	
	// Stream opens a bidirectional streaming connection
	Stream(ctx context.Context, service, method string) (ServiceStream, error)
}

// ServiceStream represents a bidirectional stream
type ServiceStream interface {
	// Send sends a message on the stream
	Send(msg interface{}) error
	
	// Receive receives a message from the stream
	Receive(msg interface{}) error
	
	// Close closes the stream
	Close() error
}

// RetryConfig defines retry parameters
type RetryConfig struct {
	MaxAttempts  int           `json:"max_attempts"`
	InitialDelay time.Duration `json:"initial_delay"`
	MaxDelay     time.Duration `json:"max_delay"`
	Multiplier   float64       `json:"multiplier"`
}

// LoadBalancer provides load balancing for service calls
type LoadBalancer interface {
	// Choose selects a service instance based on the strategy
	Choose(instances []*ServiceInstance) (*ServiceInstance, error)
	
	// Update updates the load balancer state
	Update(instances []*ServiceInstance) error
}

// LoadBalancerStrategy defines load balancing algorithms
type LoadBalancerStrategy string

const (
	LoadBalancerStrategyRoundRobin LoadBalancerStrategy = "round_robin"
	LoadBalancerStrategyRandom     LoadBalancerStrategy = "random"
	LoadBalancerStrategyLeastConn  LoadBalancerStrategy = "least_conn"
	LoadBalancerStrategyWeighted   LoadBalancerStrategy = "weighted"
)

// CircuitBreaker provides circuit breaker functionality
type CircuitBreaker interface {
	// Call executes a function with circuit breaker protection
	Call(ctx context.Context, fn func() error) error
	
	// GetState returns the current circuit breaker state
	GetState() CircuitBreakerState
	
	// Reset resets the circuit breaker
	Reset()
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState string

const (
	CircuitBreakerStateClosed    CircuitBreakerState = "closed"
	CircuitBreakerStateOpen      CircuitBreakerState = "open"
	CircuitBreakerStateHalfOpen  CircuitBreakerState = "half_open"
)

// RateLimiter provides rate limiting functionality
type RateLimiter interface {
	// Allow checks if a request is allowed
	Allow(ctx context.Context, key string) (bool, error)
	
	// AllowN checks if n requests are allowed
	AllowN(ctx context.Context, key string, n int) (bool, error)
	
	// Wait blocks until a request is allowed
	Wait(ctx context.Context, key string) error
	
	// WaitN blocks until n requests are allowed
	WaitN(ctx context.Context, key string, n int) error
}

// Authenticator provides authentication functionality
type Authenticator interface {
	// Authenticate validates credentials and returns a token
	Authenticate(ctx context.Context, credentials *Credentials) (*AuthToken, error)
	
	// Validate validates a token
	Validate(ctx context.Context, token string) (*Claims, error)
	
	// Refresh refreshes an expired token
	Refresh(ctx context.Context, refreshToken string) (*AuthToken, error)
	
	// Revoke revokes a token
	Revoke(ctx context.Context, token string) error
}

// Credentials represents authentication credentials
type Credentials struct {
	Type     string            `json:"type"` // basic, api_key, oauth
	Username string            `json:"username,omitempty"`
	Password string            `json:"password,omitempty"`
	APIKey   string            `json:"api_key,omitempty"`
	Token    string            `json:"token,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// AuthToken represents an authentication token
type AuthToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Claims represents token claims
type Claims struct {
	Subject     string            `json:"sub"`
	Issuer      string            `json:"iss"`
	Audience    []string          `json:"aud"`
	ExpiresAt   time.Time         `json:"exp"`
	IssuedAt    time.Time         `json:"iat"`
	Roles       []string          `json:"roles,omitempty"`
	Permissions []string          `json:"permissions,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Authorizer provides authorization functionality
type Authorizer interface {
	// Authorize checks if a subject has permission for an action
	Authorize(ctx context.Context, subject, resource, action string) (bool, error)
	
	// GetPermissions returns all permissions for a subject
	GetPermissions(ctx context.Context, subject string) ([]Permission, error)
	
	// GrantPermission grants a permission to a subject
	GrantPermission(ctx context.Context, subject string, permission Permission) error
	
	// RevokePermission revokes a permission from a subject
	RevokePermission(ctx context.Context, subject string, permission Permission) error
}

// Permission represents a permission
type Permission struct {
	Resource string   `json:"resource"`
	Actions  []string `json:"actions"`
}

// ServiceMesh provides service mesh capabilities
type ServiceMesh interface {
	// InjectSidecar injects a sidecar proxy
	InjectSidecar(ctx context.Context, deployment *DeploymentSpec) error
	
	// ConfigureTrafficPolicy configures traffic management
	ConfigureTrafficPolicy(ctx context.Context, policy *TrafficPolicy) error
	
	// EnableMTLS enables mutual TLS
	EnableMTLS(ctx context.Context, namespace string) error
	
	// GetMetrics retrieves service mesh metrics
	GetMetrics(ctx context.Context, service string) (*MeshMetrics, error)
}

// DeploymentSpec represents a deployment specification
type DeploymentSpec struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels"`
	Containers []Container      `json:"containers"`
}

// Container represents a container specification
type Container struct {
	Name  string            `json:"name"`
	Image string            `json:"image"`
	Ports []int             `json:"ports"`
	Env   map[string]string `json:"env,omitempty"`
}

// TrafficPolicy defines traffic management rules
type TrafficPolicy struct {
	Name        string               `json:"name"`
	Source      string               `json:"source"`
	Destination string               `json:"destination"`
	Rules       []TrafficRule        `json:"rules"`
}

// TrafficRule defines a traffic rule
type TrafficRule struct {
	Match   TrafficMatch  `json:"match,omitempty"`
	Route   []TrafficRoute `json:"route"`
	Timeout time.Duration  `json:"timeout,omitempty"`
	Retry   *RetryPolicy   `json:"retry,omitempty"`
}

// TrafficMatch defines matching criteria
type TrafficMatch struct {
	Headers map[string]string `json:"headers,omitempty"`
	Method  string            `json:"method,omitempty"`
	Path    string            `json:"path,omitempty"`
}

// TrafficRoute defines routing destination
type TrafficRoute struct {
	Destination string `json:"destination"`
	Weight      int    `json:"weight"`
}

// MeshMetrics contains service mesh metrics
type MeshMetrics struct {
	RequestRate     float64 `json:"request_rate"`
	ErrorRate       float64 `json:"error_rate"`
	P50Latency      float64 `json:"p50_latency_ms"`
	P95Latency      float64 `json:"p95_latency_ms"`
	P99Latency      float64 `json:"p99_latency_ms"`
}

// RealtimeService provides real-time updates via WebSocket
type RealtimeService interface {
	// GetHub returns the WebSocket hub for managing connections
	GetHub() *websocket.Hub
	
	// SendAgentUpdate broadcasts agent status update
	SendAgentUpdate(update websocket.AgentStatusUpdate) error
	
	// SendExperimentUpdate broadcasts experiment update
	SendExperimentUpdate(update websocket.ExperimentUpdateEvent) error
	
	// SendMetricFlow broadcasts metric flow update
	SendMetricFlow(update websocket.MetricFlowUpdate) error
	
	// SendTaskProgress broadcasts task progress update
	SendTaskProgress(update websocket.TaskProgressUpdate) error
	
	// SendAlert broadcasts an alert
	SendAlert(alert websocket.AlertEvent) error
}

// CostCalculator provides real-time cost calculation
type CostCalculator interface {
	// CalculateMetricCost calculates cost for a specific metric
	CalculateMetricCost(metric string, cardinality int64) float64
	
	// GetCostBreakdown returns cost breakdown by metric
	GetCostBreakdown() map[string]float64
	
	// GetMetricFlow returns current metric flow data
	GetMetricFlow() websocket.MetricFlowUpdate
	
	// ProjectMonthlySavings projects savings based on current vs optimized
	ProjectMonthlySavings(current, optimized float64) float64
}

// AgentRegistry manages agent status and information
type AgentRegistry interface {
	// RegisterAgent registers a new agent
	RegisterAgent(ctx context.Context, agent *Agent) error
	
	// UpdateAgentStatus updates agent status
	UpdateAgentStatus(ctx context.Context, hostID string, status websocket.AgentStatusUpdate) error
	
	// GetAgentStatus returns current agent status
	GetAgentStatus(ctx context.Context, hostID string) (*websocket.AgentStatusUpdate, error)
	
	// GetFleetStatus returns status of all agents
	GetFleetStatus(ctx context.Context) ([]websocket.AgentStatusUpdate, error)
	
	// GetAgentsByGroup returns agents in a specific group
	GetAgentsByGroup(ctx context.Context, group string) ([]websocket.AgentStatusUpdate, error)
}

// Agent represents a Phoenix agent
type Agent struct {
	HostID       string            `json:"host_id"`
	Hostname     string            `json:"hostname"`
	Group        string            `json:"group"`
	Tags         map[string]string `json:"tags"`
	Capabilities []string          `json:"capabilities"`
	Version      string            `json:"version"`
	StartedAt    time.Time         `json:"started_at"`
}