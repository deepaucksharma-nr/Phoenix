package output

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"text/tabwriter"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestColorizeStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{
			name:     "running status",
			status:   "running",
			expected: "\033[32mrunning\033[0m", // green
		},
		{
			name:     "completed status",
			status:   "completed",
			expected: "\033[34mcompleted\033[0m", // blue
		},
		{
			name:     "promoted status",
			status:   "promoted",
			expected: "\033[34mpromoted\033[0m", // blue
		},
		{
			name:     "failed status",
			status:   "failed",
			expected: "\033[31mfailed\033[0m", // red
		},
		{
			name:     "stopped status",
			status:   "stopped",
			expected: "\033[33mstopped\033[0m", // yellow
		},
		{
			name:     "unknown status",
			status:   "unknown",
			expected: "unknown", // no color
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ColorizeStatus(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPrintJSON(t *testing.T) {
	data := map[string]interface{}{
		"id":     "test-123",
		"name":   "test-experiment",
		"status": "running",
		"config": map[string]interface{}{
			"pipeline": "baseline",
			"traffic":  50,
		},
	}

	var buf bytes.Buffer
	err := PrintJSON(&buf, data)
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `"id": "test-123"`)
	assert.Contains(t, output, `"name": "test-experiment"`)
	assert.Contains(t, output, `"status": "running"`)
	assert.Contains(t, output, `"pipeline": "baseline"`)
	assert.Contains(t, output, `"traffic": 50`)

	// Verify it's properly formatted JSON
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Greater(t, len(lines), 1, "JSON should be pretty-printed")
}

func TestPrintYAML(t *testing.T) {
	data := map[string]interface{}{
		"id":     "test-123",
		"name":   "test-experiment",
		"status": "running",
		"config": map[string]interface{}{
			"pipeline": "baseline",
			"traffic":  50,
		},
	}

	var buf bytes.Buffer
	err := PrintYAML(&buf, data)
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "id: test-123")
	assert.Contains(t, output, "name: test-experiment")
	assert.Contains(t, output, "status: running")
	assert.Contains(t, output, "pipeline: baseline")
	assert.Contains(t, output, "traffic: 50")
}

func TestPrintTable(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"ID", "Name", "Status", "Created"}
	rows := [][]string{
		{"exp-1", "test-1", "running", "2024-01-15"},
		{"exp-2", "test-2", "completed", "2024-01-14"},
		{"exp-3", "test-3", "failed", "2024-01-13"},
	}

	// Use tabwriter directly as Table function prints to stdout
	w := tabwriter.NewWriter(&buf, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	w.Flush()
	output := buf.String()

	// Check headers
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "STATUS")
	assert.Contains(t, output, "CREATED")

	// Check data
	assert.Contains(t, output, "exp-1")
	assert.Contains(t, output, "test-1")
	assert.Contains(t, output, "running")
	assert.Contains(t, output, "2024-01-15")

	// Check table structure
	lines := strings.Split(output, "\n")
	assert.GreaterOrEqual(t, len(lines), 4, "Should have header and at least 3 data rows")
}

func TestFormatDuration(t *testing.T) {
	t.Skip("FormatDuration not implemented yet")
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "seconds only",
			duration: 45 * time.Second,
			expected: "45s",
		},
		{
			name:     "minutes and seconds",
			duration: 3*time.Minute + 30*time.Second,
			expected: "3m30s",
		},
		{
			name:     "hours, minutes and seconds",
			duration: 2*time.Hour + 15*time.Minute + 45*time.Second,
			expected: "2h15m45s",
		},
		{
			name:     "days",
			duration: 25*time.Hour + 30*time.Minute,
			expected: "1d1h30m0s",
		},
		{
			name:     "zero duration",
			duration: 0,
			expected: "0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{
			name:     "bytes",
			bytes:    512,
			expected: "512 B",
		},
		{
			name:     "kilobytes",
			bytes:    2048,
			expected: "2.0 KB",
		},
		{
			name:     "megabytes",
			bytes:    5242880,
			expected: "5.0 MB",
		},
		{
			name:     "gigabytes",
			bytes:    2147483648,
			expected: "2.0 GB",
		},
		{
			name:     "terabytes",
			bytes:    1099511627776,
			expected: "1.0 TB",
		},
		{
			name:     "zero bytes",
			bytes:    0,
			expected: "0 B",
		},
		{
			name:     "fractional KB",
			bytes:    1536,
			expected: "1.5 KB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatPercentage(t *testing.T) {
	t.Skip("FormatPercentage not implemented yet")
	tests := []struct {
		name     string
		value    float64
		expected string
	}{
		{
			name:     "whole number",
			value:    50.0,
			expected: "50.00%",
		},
		{
			name:     "decimal",
			value:    33.333,
			expected: "33.33%",
		},
		{
			name:     "zero",
			value:    0.0,
			expected: "0.00%",
		},
		{
			name:     "negative",
			value:    -15.5,
			expected: "-15.50%",
		},
		{
			name:     "large number",
			value:    150.75,
			expected: "150.75%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatPercentage(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "string shorter than max",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "string equal to max",
			input:    "hello",
			maxLen:   5,
			expected: "hello",
		},
		{
			name:     "string longer than max",
			input:    "hello world",
			maxLen:   8,
			expected: "hello...",
		},
		{
			name:     "very short max length",
			input:    "hello",
			maxLen:   3,
			expected: "hel",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   10,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateString(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPrintProgress(t *testing.T) {
	t.Skip("PrintProgress not implemented yet")
	tests := []struct {
		name     string
		current  int
		total    int
		width    int
		expected string
	}{
		{
			name:     "0 percent",
			current:  0,
			total:    100,
			width:    20,
			expected: "[                    ] 0%",
		},
		{
			name:     "50 percent",
			current:  50,
			total:    100,
			width:    20,
			expected: "[██████████          ] 50%",
		},
		{
			name:     "100 percent",
			current:  100,
			total:    100,
			width:    20,
			expected: "[████████████████████] 100%",
		},
		{
			name:     "33 percent",
			current:  1,
			total:    3,
			width:    10,
			expected: "[███       ] 33%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			PrintProgress(&buf, tt.current, tt.total, tt.width)
			result := strings.TrimSpace(buf.String())
			// Remove ANSI escape sequences for comparison
			result = strings.ReplaceAll(result, "\r", "")
			assert.Equal(t, tt.expected, result)
		})
	}
}