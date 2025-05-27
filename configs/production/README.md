# Phoenix Production Configuration

## Overview
Production-ready configurations for the Phoenix observability cost optimization platform.

## Phoenix Platform Features
- **70% cost reduction** in observability expenses
- **Agent-based architecture** with task polling
- **A/B testing** for safe pipeline rollouts
- **Real-time monitoring** via WebSocket

## Structure
```
production/
├── otel_collector_main_prod.yaml  # OpenTelemetry configuration
├── tls/
│   └── generate_certs.sh          # TLS certificate generation
└── README.md
```

## Production Deployment

### Prerequisites
- Kubernetes cluster or Docker Compose
- PostgreSQL database (managed or self-hosted)
- TLS certificates for secure communication
- Agent nodes with network access to control plane

### Environment Variables
```bash
# Core Configuration
DATABASE_URL=postgresql://user:pass@host:5432/phoenix
JWT_SECRET=your-secure-jwt-secret
PORT=8080

# Monitoring
PROMETHEUS_URL=http://prometheus:9090
PUSHGATEWAY_URL=http://pushgateway:9091

# Features
ENABLE_WEBSOCKET=true
ENVIRONMENT=production
```

### Security
- All agent communication over TLS
- JWT authentication for API access
- X-Agent-Host-ID header for agent auth
- Database connections encrypted

### Validation
```bash
# Validate production configuration
go run cmd/api/main.go --validate-config

# Test agent connectivity
curl -H "X-Agent-Host-ID: test-agent" http://api:8080/api/v1/agent/tasks
```