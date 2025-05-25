package simulator

import (
	"time"
)

// Profile represents a simulation profile with process patterns
type Profile struct {
	Name        string
	Patterns    []ProcessPattern
	ChurnRate   float64 // Percentage of processes to restart per hour
	ChaosConfig *ChaosConfig
}

// ProcessPattern defines how to create simulated processes
type ProcessPattern struct {
	NameTemplate string
	CPUPattern   string // steady, spiky, growing, random
	MemPattern   string // steady, spiky, growing, random
	Lifetime     time.Duration
	Count        int
	Priority     string // critical, high, medium, low
}

// ChaosConfig defines chaos engineering parameters
type ChaosConfig struct {
	FailureRate         float64 // Probability of random process failure
	CPUSpikeProbability float64 // Probability of CPU spike
	MemoryLeakRate      float64 // Rate of memory growth for leak simulation
	NetworkLatency      int     // Additional network latency in ms
}

// Default profiles for different simulation types
var profiles = map[string]*Profile{
	"realistic": {
		Name: "realistic",
		Patterns: []ProcessPattern{
			{NameTemplate: "nginx-worker-%d", CPUPattern: "steady", MemPattern: "steady", Count: 4, Priority: "high"},
			{NameTemplate: "postgres-%d", CPUPattern: "spiky", MemPattern: "growing", Count: 2, Priority: "critical"},
			{NameTemplate: "redis-server-%d", CPUPattern: "steady", MemPattern: "steady", Count: 1, Priority: "critical"},
			{NameTemplate: "python-app-%d", CPUPattern: "spiky", MemPattern: "spiky", Count: 8, Priority: "medium"},
			{NameTemplate: "node-service-%d", CPUPattern: "random", MemPattern: "steady", Count: 6, Priority: "medium"},
			{NameTemplate: "chrome-tab-%d", CPUPattern: "random", MemPattern: "growing", Lifetime: 5 * time.Minute, Count: 20, Priority: "low"},
			{NameTemplate: "cron-job-%d", CPUPattern: "spiky", MemPattern: "steady", Lifetime: 1 * time.Minute, Count: 5, Priority: "low"},
		},
		ChurnRate: 0.1, // 10% of processes restart per hour
	},
	"high-cardinality": {
		Name: "high-cardinality",
		Patterns: []ProcessPattern{
			{NameTemplate: "microservice-%d-%d", CPUPattern: "random", MemPattern: "random", Count: 100, Priority: "medium"},
			{NameTemplate: "worker-%s-%d", CPUPattern: "spiky", MemPattern: "random", Count: 50, Priority: "low"},
			{NameTemplate: "job-%s-%s-%d", CPUPattern: "random", MemPattern: "random", Lifetime: 1 * time.Minute, Count: 200, Priority: "low"},
			{NameTemplate: "container-%d", CPUPattern: "steady", MemPattern: "growing", Count: 150, Priority: "medium"},
			{NameTemplate: "sidecar-%d", CPUPattern: "steady", MemPattern: "steady", Count: 100, Priority: "low"},
			{NameTemplate: "agent-%d", CPUPattern: "spiky", MemPattern: "steady", Count: 50, Priority: "high"},
		},
		ChurnRate: 0.5, // 50% churn rate
	},
	"process-churn": {
		Name: "process-churn",
		Patterns: []ProcessPattern{
			{NameTemplate: "short-lived-%d", CPUPattern: "spiky", MemPattern: "steady", Lifetime: 30 * time.Second, Count: 50, Priority: "low"},
			{NameTemplate: "batch-job-%d", CPUPattern: "steady", MemPattern: "growing", Lifetime: 2 * time.Minute, Count: 30, Priority: "medium"},
			{NameTemplate: "temp-worker-%d", CPUPattern: "random", MemPattern: "random", Lifetime: 1 * time.Minute, Count: 40, Priority: "low"},
			{NameTemplate: "lambda-function-%d", CPUPattern: "spiky", MemPattern: "steady", Lifetime: 15 * time.Second, Count: 100, Priority: "low"},
			{NameTemplate: "ci-runner-%d", CPUPattern: "steady", MemPattern: "growing", Lifetime: 5 * time.Minute, Count: 20, Priority: "medium"},
		},
		ChurnRate: 0.8, // 80% churn rate
	},
}

// MetricsEmitter defines the interface for emitting process metrics
type MetricsEmitter interface {
	EmitProcessMetrics(process *SimulatedProcess)
	EmitSystemMetrics(totalCPU, totalMemory float64, processCount int)
}

// ProcessClassifier classifies processes by priority
type ProcessClassifier struct {
	Rules []ClassificationRule
}

// ClassificationRule defines how to classify a process
type ClassificationRule struct {
	Pattern  string
	Priority string
}

// DefaultClassificationRules provides default process classification
var DefaultClassificationRules = []ClassificationRule{
	{Pattern: "postgres", Priority: "critical"},
	{Pattern: "redis", Priority: "critical"},
	{Pattern: "mysql", Priority: "critical"},
	{Pattern: "nginx", Priority: "high"},
	{Pattern: "apache", Priority: "high"},
	{Pattern: "node", Priority: "medium"},
	{Pattern: "python", Priority: "medium"},
	{Pattern: "java", Priority: "medium"},
	{Pattern: "chrome", Priority: "low"},
	{Pattern: "temp", Priority: "low"},
	{Pattern: "job", Priority: "low"},
}

// SimulatorConfig holds configuration for the process simulator
type SimulatorConfig struct {
	MetricsEndpoint     string
	PrometheusPort      int
	EnableChaos         bool
	EnableMetrics       bool
	ClassificationRules []ClassificationRule
}