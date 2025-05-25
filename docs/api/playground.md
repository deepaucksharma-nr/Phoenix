# API Playground

Try out the Phoenix Platform APIs interactively using our API playground.

## REST API Explorer

<div id="swagger-ui"></div>

<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.9.0/swagger-ui.css">
<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>

<script>
window.onload = function() {
  const ui = SwaggerUIBundle({
    url: "/assets/openapi.yaml",
    dom_id: '#swagger-ui',
    deepLinking: true,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl
    ],
    layout: "StandaloneLayout",
    tryItOutEnabled: true,
    supportedSubmitMethods: ['get', 'post', 'put', 'delete', 'patch'],
    onComplete: function() {
      // Custom styling to match Material theme
      const swaggerContainer = document.querySelector('.swagger-ui');
      if (swaggerContainer) {
        swaggerContainer.style.fontFamily = 'var(--md-text-font-family)';
      }
    }
  });
  window.ui = ui;
}
</script>

<style>
/* Swagger UI Theme Overrides */
.swagger-ui .topbar {
  display: none;
}

.swagger-ui .info .title {
  color: var(--md-primary-fg-color);
}

.swagger-ui .btn.authorize {
  background-color: var(--md-primary-fg-color);
  color: white;
}

.swagger-ui .btn.execute {
  background-color: var(--md-accent-fg-color);
  color: white;
}

.swagger-ui select, 
.swagger-ui input[type=text], 
.swagger-ui textarea {
  background: var(--md-code-bg-color);
  color: var(--md-code-fg-color);
  border: 1px solid var(--md-default-fg-color--lightest);
}

.swagger-ui .scheme-container {
  background: var(--md-code-bg-color);
  border: 1px solid var(--md-default-fg-color--lightest);
}

.swagger-ui .opblock.opblock-get .opblock-summary-method {
  background: #61affe;
}

.swagger-ui .opblock.opblock-post .opblock-summary-method {
  background: #49cc90;
}

.swagger-ui .opblock.opblock-put .opblock-summary-method {
  background: #fca130;
}

.swagger-ui .opblock.opblock-delete .opblock-summary-method {
  background: #f93e3e;
}

.swagger-ui .opblock.opblock-patch .opblock-summary-method {
  background: #50e3c2;
}
</style>

## Authentication

To use the API playground:

1. Click the **Authorize** button above
2. Enter your API key or JWT token
3. Click **Authorize** to apply authentication to all requests

## Available Endpoints

The playground includes all Phoenix Platform API endpoints:

- **Experiments** - Create and manage A/B testing experiments
- **Pipelines** - Configure OpenTelemetry pipelines
- **Metrics** - Query metrics and analytics
- **Load Simulations** - Manage load testing scenarios

## Example Requests

### Create an Experiment

1. Navigate to `POST /v1/experiments`
2. Click **Try it out**
3. Modify the request body:
   ```json
   {
     "name": "my-test-experiment",
     "baseline_pipeline": "process-baseline-v1",
     "candidate_pipeline": "process-topk-v1",
     "duration": "24h"
   }
   ```
4. Click **Execute**

### Query Metrics

1. Navigate to `GET /v1/experiments/{id}/metrics`
2. Enter an experiment ID
3. Set query parameters (optional):
   - `start`: Start time
   - `end`: End time
   - `resolution`: Data resolution
4. Click **Execute**

## Response Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request - Check your parameters |
| 401 | Unauthorized - Check your authentication |
| 404 | Not Found - Resource doesn't exist |
| 429 | Rate Limited - Too many requests |
| 500 | Server Error - Contact support |

## Rate Limits

API calls from the playground are subject to the same rate limits as regular API calls:

- **Authenticated**: 1000 requests/hour
- **Unauthenticated**: 100 requests/hour

## Need Help?

- Check the [API Reference](./rest.md) for detailed documentation
- Join our [Slack community](https://phoenix-community.slack.com)
- Report issues on [GitHub](https://github.com/phoenix-platform/phoenix/issues)