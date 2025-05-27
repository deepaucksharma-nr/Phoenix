# Phoenix Platform Configuration Reference

## Overview

This document provides a complete reference for all configuration options across Phoenix Platform components. Configuration can be provided through environment variables, configuration files, or command-line flags.

## Configuration Precedence

1. Command-line flags (highest priority)
2. Environment variables
3. Configuration files
4. Default values (lowest priority)

## Phoenix API Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `PHOENIX_API_PORT` | API server port | `8080` | No |
| `PHOENIX_API_HOST` | API server host | `0.0.0.0` | No |
| `PHOENIX_DB_HOST` | PostgreSQL host | `localhost` | Yes |
| `PHOENIX_DB_PORT` | PostgreSQL port | `5432` | No |
| `PHOENIX_DB_NAME` | Database name | `phoenix` | Yes |
| `PHOENIX_DB_USER` | Database user | `phoenix` | Yes |
| `PHOENIX_DB_PASSWORD` | Database password | - | Yes |
| `PHOENIX_DB_SSLMODE` | SSL mode | `require` | No |
| `PHOENIX_JWT_SECRET` | JWT signing secret | - | Yes |
| `PHOENIX_JWT_EXPIRY` | Token expiry duration | `24h` | No |
| `PHOENIX_LOG_LEVEL` | Log level | `info` | No |
| `PHOENIX_LOG_FORMAT` | Log format (json/text) | `json` | No |
| `PHOENIX_CORS_ORIGINS` | Allowed CORS origins | `*` | No |
| `PHOENIX_RATE_LIMIT` | Rate limit per minute | `1000` | No |
| `PHOENIX_WS_ENABLED` | Enable WebSocket | `true` | No |
| `PHOENIX_WS_MAX_CONN` | Max WS connections | `10000` | No |
| `PHOENIX_METRICS_ENABLED` | Enable Prometheus metrics | `true` | No |
| `PHOENIX_METRICS_PORT` | Metrics port | `9090` | No |

### Configuration File (config.yaml)

```yaml
# Phoenix API Configuration
api:
  host: 0.0.0.0
  port: 8080
  read_timeout: 30s
  write_timeout: 30s
  shutdown_timeout: 10s

database:
  host: localhost
  port: 5432
  name: phoenix
  user: phoenix
  password: ${PHOENIX_DB_PASSWORD}
  max_connections: 100
  max_idle_connections: 10
  connection_max_lifetime: 1h
  ssl_mode: require

auth:
  jwt_secret: ${PHOENIX_JWT_SECRET}
  jwt_expiry: 24h
  refresh_expiry: 168h
  bcrypt_cost: 10

websocket:
  enabled: true
  max_connections: 10000
  ping_interval: 30s
  pong_timeout: 60s
  write_buffer_size: 1024
  read_buffer_size: 1024

task_queue:
  poll_timeout: 30s
  max_retries: 3
  retry_delay: 5s
  cleanup_interval: 1h
  task_retention: 168h

metrics:
  enabled: true
  port: 9090
  path: /metrics
  namespace: phoenix_api

logging:
  level: info
  format: json
  output: stdout
  file:
    enabled: false
    path: /var/log/phoenix/api.log
    max_size: 100
    max_age: 7
    max_backups: 5

cors:
  allowed_origins:
    - http://localhost:3000
    - https://phoenix.example.com
  allowed_methods:
    - GET
    - POST
    - PUT
    - DELETE
    - OPTIONS
  allowed_headers:
    - Authorization
    - Content-Type
    - X-Request-ID
  expose_headers:
    - X-Request-ID
  max_age: 3600

rate_limiting:
  enabled: true
  requests_per_minute: 1000
  burst: 100
  exclude_paths:
    - /health
    - /metrics
```

## Phoenix Agent Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `PHOENIX_AGENT_ID` | Unique agent identifier | hostname | No |
| `PHOENIX_API_URL` | Phoenix API URL | `http://localhost:8080` | Yes |
| `PHOENIX_AGENT_LABELS` | Agent labels (key=value) | - | No |
| `PHOENIX_POLL_INTERVAL` | Task poll interval | `10s` | No |
| `PHOENIX_HEARTBEAT_INTERVAL` | Heartbeat interval | `30s` | No |
| `PHOENIX_METRICS_INTERVAL` | Metrics report interval | `60s` | No |
| `PHOENIX_OTEL_CONFIG_PATH` | OTel config directory | `/etc/otel` | No |
| `PHOENIX_OTEL_BINARY` | OTel collector binary | `otelcol` | No |
| `PHOENIX_LOG_LEVEL` | Log level | `info` | No |
| `PHOENIX_WORK_DIR` | Working directory | `/var/lib/phoenix` | No |

### Configuration File (agent.yaml)

```yaml
# Phoenix Agent Configuration
agent:
  id: ${HOSTNAME}
  labels:
    region: us-east-1
    environment: production
    cluster: web-tier

api:
  url: http://phoenix-api:8080
  timeout: 30s
  retry_max: 3
  retry_delay: 5s

polling:
  task_interval: 10s
  heartbeat_interval: 30s
  metrics_interval: 60s
  jitter: 5s

otel:
  config_dir: /etc/otel
  binary_path: /usr/bin/otelcol
  restart_delay: 10s
  health_check_port: 13133
  graceful_shutdown: 30s

pipelines:
  validate_before_deploy: true
  rollback_on_error: true
  max_rollback_attempts: 3

metrics:
  buffer_size: 10000
  flush_interval: 30s
  include_system_metrics: true
  system_metrics_interval: 60s

logging:
  level: info
  format: json
  output: stdout

storage:
  work_dir: /var/lib/phoenix
  config_cache_dir: /var/cache/phoenix
  cleanup_interval: 24h
  max_cache_size: 1GB
```

## Phoenix CLI Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `PHOENIX_API_URL` | API endpoint | `http://localhost:8080` | No |
| `PHOENIX_CONFIG_DIR` | Config directory | `~/.phoenix` | No |
| `PHOENIX_OUTPUT_FORMAT` | Output format | `table` | No |
| `PHOENIX_NO_COLOR` | Disable colors | `false` | No |
| `PHOENIX_TIMEOUT` | Request timeout | `30s` | No |

### Configuration File (~/.phoenix/config.yaml)

```yaml
# Phoenix CLI Configuration
api:
  url: http://localhost:8080
  timeout: 30s

auth:
  token_file: ~/.phoenix/token
  auto_refresh: true

output:
  format: table  # table, json, yaml
  color: true
  truncate: true
  max_width: 120

contexts:
  default:
    api_url: http://localhost:8080
    timeout: 30s
  
  production:
    api_url: https://phoenix.example.com
    timeout: 60s
  
  staging:
    api_url: https://phoenix-staging.example.com
    timeout: 30s

current_context: default

aliases:
  exp: experiment
  pipe: pipeline
  dep: deployment
```

## Dashboard Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `VITE_API_URL` | API base URL | `http://localhost:8080` | Yes |
| `VITE_WS_URL` | WebSocket URL | `ws://localhost:8080` | Yes |
| `VITE_REFRESH_INTERVAL` | Data refresh interval | `5000` | No |
| `VITE_THEME` | Default theme | `light` | No |
| `VITE_ENABLE_ANALYTICS` | Enable analytics | `false` | No |

### Build Configuration (vite.config.ts)

```typescript
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: process.env.VITE_API_URL || 'http://localhost:8080',
        changeOrigin: true,
      },
      '/ws': {
        target: process.env.VITE_WS_URL || 'ws://localhost:8080',
        ws: true,
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
  },
});
```

## Pipeline Configuration

### Adaptive Filter Pipeline

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  adaptive_filter:
    # Importance threshold (0.0-1.0)
    importance_threshold: 0.7
    
    # Evaluation interval
    evaluation_interval: 5m
    
    # Minimum samples before filtering
    min_sample_size: 1000
    
    # Namespace filters
    namespace_filters:
      include: ["app_*", "business_*"]
      exclude: ["test_*", "debug_*"]
    
    # Preservation rules
    always_keep:
      - name: "critical_business_metric"
      - labels: 
          severity: "critical"
    
    # ML model settings
    model:
      type: "gradient_boost"
      update_interval: 1h
      feature_importance_method: "shap"

exporters:
  prometheus:
    endpoint: 0.0.0.0:8889
    namespace: optimized
    send_timestamps: true
    metric_expiration: 5m

service:
  pipelines:
    metrics:
      receivers: [otlp]
      processors: [adaptive_filter]
      exporters: [prometheus]
```

### TopK Pipeline

```yaml
processors:
  topk:
    # Number of top metrics to keep
    k: 10000
    
    # Ranking method
    ranking_method: "frequency"  # frequency, value, composite
    
    # Time window for ranking
    window_duration: 5m
    
    # Update interval
    update_interval: 1m
    
    # Cache settings
    cache:
      size: 50000
      ttl: 10m
    
    # Grouping rules
    group_by:
      - "namespace"
      - "job"
```

## Deployment Configuration

### Docker Compose Configuration

```yaml
# Configuration via environment files
# .env.production
PHOENIX_DB_HOST=postgres
PHOENIX_DB_PORT=5432
PHOENIX_DB_NAME=phoenix
PHOENIX_DB_USER=phoenix
PHOENIX_DB_PASSWORD=secure_password
PHOENIX_JWT_SECRET=your_jwt_secret
PHOENIX_API_PORT=8080
PHOENIX_LOG_LEVEL=info
```

### Docker Compose Environment

```yaml
# docker-compose.yml
services:
  phoenix-api:
    environment:
      - PHOENIX_DB_HOST=postgres
      - PHOENIX_DB_PASSWORD=${POSTGRES_PASSWORD}
      - PHOENIX_JWT_SECRET=${JWT_SECRET}
      - PHOENIX_LOG_LEVEL=debug
    env_file:
      - .env.local
```

## Security Configuration

### TLS/SSL Settings

```yaml
# TLS configuration for production
tls:
  enabled: true
  cert_file: /etc/phoenix/tls/server.crt
  key_file: /etc/phoenix/tls/server.key
  ca_file: /etc/phoenix/tls/ca.crt
  client_auth: require
  min_version: "1.2"
  cipher_suites:
    - TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
    - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
```

### Authentication Providers

```yaml
auth:
  providers:
    - type: local
      enabled: true
    
    - type: ldap
      enabled: true
      url: ldaps://ldap.example.com:636
      bind_dn: "cn=phoenix,ou=services,dc=example,dc=com"
      bind_password: ${LDAP_BIND_PASSWORD}
      search_base: "ou=users,dc=example,dc=com"
      search_filter: "(uid={{username}})"
    
    - type: oauth2
      enabled: true
      provider: google
      client_id: ${OAUTH_CLIENT_ID}
      client_secret: ${OAUTH_CLIENT_SECRET}
      redirect_url: https://phoenix.example.com/auth/callback
```

## Monitoring Configuration

### Prometheus Scrape Config

```yaml
scrape_configs:
  - job_name: 'phoenix-api'
    static_configs:
      - targets: ['phoenix-api:9090']
    metric_relabel_configs:
      - source_labels: [__name__]
        regex: 'go_.*'
        action: drop

  - job_name: 'phoenix-agents'
    file_sd_configs:
      - files:
          - '/etc/prometheus/targets/agents.yml'
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance
```

### Alerting Rules

```yaml
groups:
  - name: phoenix_alerts
    rules:
      - alert: HighCardinalityDetected
        expr: phoenix_pipeline_cardinality > 1000000
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High cardinality detected"
          description: "Pipeline {{ $labels.pipeline }} has cardinality > 1M"
      
      - alert: ExperimentFailed
        expr: phoenix_experiment_status == 3
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Experiment failed"
          description: "Experiment {{ $labels.experiment_id }} has failed"
```

## Performance Tuning

### Database Connection Pool

```yaml
database:
  # Connection pool settings
  max_connections: 100
  max_idle_connections: 10
  connection_max_lifetime: 1h
  
  # Query performance
  statement_timeout: 30s
  lock_timeout: 10s
  
  # Maintenance
  auto_vacuum: true
  vacuum_interval: 24h
```

### Resource Limits

```yaml
resources:
  api:
    requests:
      cpu: 500m
      memory: 512Mi
    limits:
      cpu: 2000m
      memory: 2Gi
  
  agent:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 512Mi
```

## Troubleshooting

### Debug Mode

Enable debug logging:
```bash
export PHOENIX_LOG_LEVEL=debug
export PHOENIX_LOG_FORMAT=text
```

### Trace Sampling

```yaml
tracing:
  enabled: true
  sampling_rate: 0.1
  exporter: jaeger
  endpoint: http://jaeger:14268/api/traces
```