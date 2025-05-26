//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPipelineDeploymentAndExperiment ensures a pipeline can be deployed
// and then referenced by a new experiment via the API.
func TestPipelineDeploymentAndExperiment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	apiURL := "http://localhost:8080"

	// Deploy pipeline
	dep := map[string]interface{}{
		"name":      fmt.Sprintf("int-dep-%d", time.Now().Unix()),
		"namespace": "default",
		"template":  "process-baseline-v1",
	}
	body, _ := json.Marshal(dep)
	resp, err := http.Post(apiURL+"/api/v1/pipeline-deployments", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Create experiment referencing the deployment
	exp := map[string]interface{}{
		"name":               "int-exp",
		"baseline_pipeline":  dep["template"],
		"candidate_pipeline": "process-intelligent-v1",
		"target_namespaces":  []string{"default"},
	}
	expBody, _ := json.Marshal(exp)
	resp, err = http.Post(apiURL+"/api/v1/experiments", "application/json", bytes.NewBuffer(expBody))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}
