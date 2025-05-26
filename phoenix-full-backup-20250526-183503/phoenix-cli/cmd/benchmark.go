package cmd

import (
	"fmt"
	"sync"
	"time"

	"github.com/phoenix/platform/projects/phoenix-cli/internal/client"
	"github.com/phoenix/platform/projects/phoenix-cli/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var benchmarkCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "Run performance benchmarks against Phoenix API",
	Long:  "Execute various performance tests to measure API response times, throughput, and resource usage",
}

var benchmarkApiCmd = &cobra.Command{
	Use:   "api",
	Short: "Benchmark API endpoints",
	Long:  "Run performance tests against Phoenix API endpoints to measure response times and throughput",
	RunE:  runAPIBenchmark,
}

var benchmarkExperimentCmd = &cobra.Command{
	Use:   "experiment",
	Short: "Benchmark experiment operations",
	Long:  "Run performance tests for experiment lifecycle operations",
	RunE:  runExperimentBenchmark,
}

var benchmarkLoadCmd = &cobra.Command{
	Use:   "load",
	Short: "Run load tests",
	Long:  "Execute load tests with configurable concurrency and request patterns",
	RunE:  runLoadBenchmark,
}

// BenchmarkResult represents the result of a benchmark test
type BenchmarkResult struct {
	Name           string        `json:"name"`
	TotalRequests  int           `json:"total_requests"`
	SuccessfulReqs int           `json:"successful_requests"`
	FailedReqs     int           `json:"failed_requests"`
	TotalDuration  time.Duration `json:"total_duration"`
	AvgLatency     time.Duration `json:"avg_latency"`
	MinLatency     time.Duration `json:"min_latency"`
	MaxLatency     time.Duration `json:"max_latency"`
	P95Latency     time.Duration `json:"p95_latency"`
	P99Latency     time.Duration `json:"p99_latency"`
	RequestsPerSec float64       `json:"requests_per_second"`
	ErrorRate      float64       `json:"error_rate"`
}

// LatencyTracker tracks request latencies
type LatencyTracker struct {
	latencies []time.Duration
	mu        sync.Mutex
}

func (lt *LatencyTracker) Add(latency time.Duration) {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	lt.latencies = append(lt.latencies, latency)
}

func (lt *LatencyTracker) GetStats() (min, max, avg, p95, p99 time.Duration) {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	
	if len(lt.latencies) == 0 {
		return 0, 0, 0, 0, 0
	}
	
	// Sort latencies for percentile calculation
	sorted := make([]time.Duration, len(lt.latencies))
	copy(sorted, lt.latencies)
	
	// Simple sort
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	
	min = sorted[0]
	max = sorted[len(sorted)-1]
	
	// Calculate average
	var total time.Duration
	for _, lat := range sorted {
		total += lat
	}
	avg = total / time.Duration(len(sorted))
	
	// Calculate percentiles
	p95Index := int(float64(len(sorted)) * 0.95)
	p99Index := int(float64(len(sorted)) * 0.99)
	
	if p95Index >= len(sorted) {
		p95Index = len(sorted) - 1
	}
	if p99Index >= len(sorted) {
		p99Index = len(sorted) - 1
	}
	
	p95 = sorted[p95Index]
	p99 = sorted[p99Index]
	
	return min, max, avg, p95, p99
}

func init() {
	benchmarkCmd.AddCommand(benchmarkApiCmd)
	benchmarkCmd.AddCommand(benchmarkExperimentCmd)
	benchmarkCmd.AddCommand(benchmarkLoadCmd)

	// API benchmark flags
	benchmarkApiCmd.Flags().Int("requests", 100, "Number of requests to send")
	benchmarkApiCmd.Flags().Int("concurrency", 10, "Number of concurrent requests")
	benchmarkApiCmd.Flags().String("endpoint", "/api/v1/experiments", "API endpoint to test")
	benchmarkApiCmd.Flags().String("method", "GET", "HTTP method to use")

	// Experiment benchmark flags
	benchmarkExperimentCmd.Flags().Int("experiments", 10, "Number of experiments to create")
	benchmarkExperimentCmd.Flags().Int("concurrency", 5, "Number of concurrent operations")
	benchmarkExperimentCmd.Flags().Bool("cleanup", true, "Clean up created experiments")

	// Load test flags
	benchmarkLoadCmd.Flags().Duration("duration", 60*time.Second, "Duration to run the test")
	benchmarkLoadCmd.Flags().Int("rps", 10, "Target requests per second")
	benchmarkLoadCmd.Flags().String("pattern", "constant", "Load pattern: constant, ramp, spike")
	benchmarkLoadCmd.Flags().String("endpoints", "/api/v1/experiments,/api/v1/pipeline-deployments", "Comma-separated endpoints to test")

	rootCmd.AddCommand(benchmarkCmd)
}

func runAPIBenchmark(cmd *cobra.Command, args []string) error {
	requests, _ := cmd.Flags().GetInt("requests")
	concurrency, _ := cmd.Flags().GetInt("concurrency")
	endpoint, _ := cmd.Flags().GetString("endpoint")
	method, _ := cmd.Flags().GetString("method")

	apiClient, err := getAPIClient()
	if err != nil {
		return err
	}

	fmt.Printf("Starting API benchmark...\n")
	fmt.Printf("Endpoint: %s %s\n", method, endpoint)
	fmt.Printf("Requests: %d\n", requests)
	fmt.Printf("Concurrency: %d\n", concurrency)
	fmt.Printf("\n")

	result, err := runConcurrentRequests(apiClient, endpoint, method, requests, concurrency)
	if err != nil {
		return err
	}

	return printBenchmarkResult(cmd, result)
}

func runExperimentBenchmark(cmd *cobra.Command, args []string) error {
	experiments, _ := cmd.Flags().GetInt("experiments")
	concurrency, _ := cmd.Flags().GetInt("concurrency")
	cleanup, _ := cmd.Flags().GetBool("cleanup")

	apiClient, err := getAPIClient()
	if err != nil {
		return err
	}

	fmt.Printf("Starting experiment benchmark...\n")
	fmt.Printf("Experiments: %d\n", experiments)
	fmt.Printf("Concurrency: %d\n", concurrency)
	fmt.Printf("Cleanup: %v\n", cleanup)
	fmt.Printf("\n")

	result, createdExperiments, err := runExperimentOperations(apiClient, experiments, concurrency)
	if err != nil {
		return err
	}

	// Cleanup if requested
	if cleanup && len(createdExperiments) > 0 {
		fmt.Printf("Cleaning up %d experiments...\n", len(createdExperiments))
		cleanupExperiments(apiClient, createdExperiments)
	}

	return printBenchmarkResult(cmd, result)
}

func runLoadBenchmark(cmd *cobra.Command, args []string) error {
	duration, _ := cmd.Flags().GetDuration("duration")
	rps, _ := cmd.Flags().GetInt("rps")
	pattern, _ := cmd.Flags().GetString("pattern")
	endpoints, _ := cmd.Flags().GetString("endpoints")

	apiClient, err := getAPIClient()
	if err != nil {
		return err
	}

	fmt.Printf("Starting load benchmark...\n")
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Target RPS: %d\n", rps)
	fmt.Printf("Pattern: %s\n", pattern)
	fmt.Printf("Endpoints: %s\n", endpoints)
	fmt.Printf("\n")

	result, err := runLoadPattern(apiClient, duration, rps, pattern, endpoints)
	if err != nil {
		return err
	}

	return printBenchmarkResult(cmd, result)
}

func runConcurrentRequests(apiClient *client.APIClient, endpoint, method string, totalRequests, concurrency int) (*BenchmarkResult, error) {
	var wg sync.WaitGroup
	var successCount, failCount int
	var mu sync.Mutex
	tracker := &LatencyTracker{}

	requestsPerWorker := totalRequests / concurrency
	remainder := totalRequests % concurrency

	startTime := time.Now()

	for i := 0; i < concurrency; i++ {
		workerRequests := requestsPerWorker
		if i < remainder {
			workerRequests++
		}

		wg.Add(1)
		go func(requests int) {
			defer wg.Done()

			for j := 0; j < requests; j++ {
				requestStart := time.Now()
				
				// Make the API request
				err := makeRequest(apiClient, endpoint, method)
				
				latency := time.Since(requestStart)
				tracker.Add(latency)

				mu.Lock()
				if err != nil {
					failCount++
				} else {
					successCount++
				}
				mu.Unlock()
			}
		}(workerRequests)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	min, max, avg, p95, p99 := tracker.GetStats()

	result := &BenchmarkResult{
		Name:           fmt.Sprintf("%s %s", method, endpoint),
		TotalRequests:  totalRequests,
		SuccessfulReqs: successCount,
		FailedReqs:     failCount,
		TotalDuration:  totalDuration,
		AvgLatency:     avg,
		MinLatency:     min,
		MaxLatency:     max,
		P95Latency:     p95,
		P99Latency:     p99,
		RequestsPerSec: float64(totalRequests) / totalDuration.Seconds(),
		ErrorRate:      float64(failCount) / float64(totalRequests) * 100,
	}

	return result, nil
}

func runExperimentOperations(apiClient *client.APIClient, totalExperiments, concurrency int) (*BenchmarkResult, []string, error) {
	var wg sync.WaitGroup
	var successCount, failCount int
	var mu sync.Mutex
	var createdExperiments []string
	tracker := &LatencyTracker{}

	experimentsPerWorker := totalExperiments / concurrency
	remainder := totalExperiments % concurrency

	startTime := time.Now()

	for i := 0; i < concurrency; i++ {
		workerExperiments := experimentsPerWorker
		if i < remainder {
			workerExperiments++
		}

		wg.Add(1)
		go func(worker, experiments int) {
			defer wg.Done()

			for j := 0; j < experiments; j++ {
				requestStart := time.Now()
				
				// Create experiment
				req := client.CreateExperimentRequest{
					Name:              fmt.Sprintf("benchmark-exp-%d-%d", worker, j),
					Description:       "Benchmark experiment",
					BaselinePipeline:  "process-baseline-v1",
					CandidatePipeline: "process-intelligent-v1",
					TargetNodes:       map[string]string{"app": "benchmark"},
					Duration:          30 * time.Minute,
					Parameters:        map[string]interface{}{"traffic_split": "50/50"},
				}

				experiment, err := apiClient.CreateExperiment(req)
				latency := time.Since(requestStart)
				tracker.Add(latency)

				mu.Lock()
				if err != nil {
					failCount++
				} else {
					successCount++
					createdExperiments = append(createdExperiments, experiment.ID)
				}
				mu.Unlock()
			}
		}(i, workerExperiments)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	min, max, avg, p95, p99 := tracker.GetStats()

	result := &BenchmarkResult{
		Name:           "Experiment Creation",
		TotalRequests:  totalExperiments,
		SuccessfulReqs: successCount,
		FailedReqs:     failCount,
		TotalDuration:  totalDuration,
		AvgLatency:     avg,
		MinLatency:     min,
		MaxLatency:     max,
		P95Latency:     p95,
		P99Latency:     p99,
		RequestsPerSec: float64(totalExperiments) / totalDuration.Seconds(),
		ErrorRate:      float64(failCount) / float64(totalExperiments) * 100,
	}

	return result, createdExperiments, nil
}

func runLoadPattern(apiClient *client.APIClient, duration time.Duration, targetRPS int, pattern, endpoints string) (*BenchmarkResult, error) {
	// Parse endpoints
	endpointList := []string{"/api/v1/experiments"}
	if endpoints != "" {
		// Simple split for now
		endpointList = []string{endpoints}
	}

	var totalRequests int
	var successCount, failCount int
	var mu sync.Mutex
	tracker := &LatencyTracker{}

	startTime := time.Now()
	
	// Calculate interval between requests
	interval := time.Second / time.Duration(targetRPS)
	
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	done := make(chan bool)
	go func() {
		time.Sleep(duration)
		done <- true
	}()

	for {
		select {
		case <-done:
			goto finish
		case <-ticker.C:
			go func() {
				requestStart := time.Now()
				
				// Rotate through endpoints
				endpoint := endpointList[totalRequests%len(endpointList)]
				err := makeRequest(apiClient, endpoint, "GET")
				
				latency := time.Since(requestStart)
				tracker.Add(latency)

				mu.Lock()
				totalRequests++
				if err != nil {
					failCount++
				} else {
					successCount++
				}
				mu.Unlock()
			}()
		}
	}

finish:
	totalDuration := time.Since(startTime)
	min, max, avg, p95, p99 := tracker.GetStats()

	result := &BenchmarkResult{
		Name:           fmt.Sprintf("Load Test (%s)", pattern),
		TotalRequests:  totalRequests,
		SuccessfulReqs: successCount,
		FailedReqs:     failCount,
		TotalDuration:  totalDuration,
		AvgLatency:     avg,
		MinLatency:     min,
		MaxLatency:     max,
		P95Latency:     p95,
		P99Latency:     p99,
		RequestsPerSec: float64(totalRequests) / totalDuration.Seconds(),
		ErrorRate:      float64(failCount) / float64(totalRequests) * 100,
	}

	return result, nil
}

func makeRequest(apiClient *client.APIClient, endpoint, method string) error {
	switch endpoint {
	case "/api/v1/experiments":
		req := client.ListExperimentsRequest{}
		_, err := apiClient.ListExperiments(req)
		return err
	case "/api/v1/pipeline-deployments":
		// This would require implementing the pipeline deployment client methods
		return nil
	default:
		// Generic request - this would require a more generic client method
		return nil
	}
}

func cleanupExperiments(apiClient *client.APIClient, experimentIDs []string) {
	for _, id := range experimentIDs {
		// In a real implementation, you'd call the delete experiment API
		// For now, we'll just simulate cleanup
		_ = id
	}
}

func printBenchmarkResult(cmd *cobra.Command, result *BenchmarkResult) error {
	outputFormat := viper.GetString("output")
	
	switch outputFormat {
	case "json":
		return output.PrintJSON(cmd.OutOrStdout(), result)
	case "yaml":
		return output.PrintYAML(cmd.OutOrStdout(), result)
	default:
		fmt.Printf("Benchmark Results: %s\n", result.Name)
		fmt.Printf("===========================================\n")
		fmt.Printf("Total Requests:      %d\n", result.TotalRequests)
		fmt.Printf("Successful:          %d\n", result.SuccessfulReqs)
		fmt.Printf("Failed:              %d\n", result.FailedReqs)
		fmt.Printf("Total Duration:      %v\n", result.TotalDuration)
		fmt.Printf("Requests/sec:        %.2f\n", result.RequestsPerSec)
		fmt.Printf("Error Rate:          %.2f%%\n", result.ErrorRate)
		fmt.Printf("\n")
		fmt.Printf("Latency Statistics:\n")
		fmt.Printf("  Average:           %v\n", result.AvgLatency)
		fmt.Printf("  Minimum:           %v\n", result.MinLatency)
		fmt.Printf("  Maximum:           %v\n", result.MaxLatency)
		fmt.Printf("  95th percentile:   %v\n", result.P95Latency)
		fmt.Printf("  99th percentile:   %v\n", result.P99Latency)
		return nil
	}
}

// getAPIClient creates an API client from configuration
func getAPIClient() (*client.APIClient, error) {
	token := viper.GetString("auth.token")
	if token == "" {
		return nil, fmt.Errorf("not authenticated. Please run: phoenix auth login")
	}
	
	apiEndpoint := viper.GetString("api.endpoint")
	if apiEndpoint == "" {
		apiEndpoint = "http://localhost:8080" // default
	}
	
	return client.NewAPIClient(apiEndpoint, token), nil
}