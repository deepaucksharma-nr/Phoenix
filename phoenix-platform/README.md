# Phoenix Process Metrics Optimization Platform

Phoenix is an automated, dashboard-driven solution for optimizing process metrics collection in New Relic Infrastructure monitoring. The platform reduces telemetry costs by 50-80% while maintaining 100% visibility for critical processes.

## 🚀 Key Features

- **Visual Pipeline Builder**: Drag-and-drop interface for creating OpenTelemetry configurations
- **Automated A/B Testing**: Side-by-side comparison of optimization strategies
- **Zero-Touch Deployment**: Automatic generation and deployment of all components
- **Intelligent Analysis**: Real-time cost and performance analytics with recommendations

## 📋 Prerequisites

- Kubernetes 1.28+
- Helm 3.12+
- Go 1.21+ (for development)
- Node.js 18+ (for dashboard development)
- New Relic account with OTLP endpoint access

## 🏗️ Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Web Dashboard  │────▶│   API Gateway   │────▶│ Experiment API  │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                                                          │
                                ┌─────────────────────────┴───────────────┐
                                │                                         │
                        ┌───────▼────────┐                      ┌────────▼────────┐
                        │  Config Gen    │                      │ Experiment Ctrl │
                        └───────┬────────┘                      └────────┬────────┘
                                │                                         │
                        ┌───────▼────────┐                      ┌────────▼────────┐
                        │   Git Repo     │◀─────────────────────│  Kubernetes API │
                        └───────┬────────┘                      └─────────────────┘
                                │                                         │
                        ┌───────▼────────┐     ┌─────────────────────────┘
                        │    ArgoCD      │     │
                        └───────┬────────┘     │
                                │              │
                        ┌───────▼──────────────▼──┐
                        │   OTel Collectors       │
                        │  (Baseline & Candidate) │
                        └───────┬─────────────────┘
                                │
                        ┌───────▼────────┐
                        │  New Relic &   │
                        │  Prometheus    │
                        └────────────────┘
```

## 🚀 Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/your-org/phoenix-platform.git
cd phoenix-platform
```

### 2. Install Phoenix using Helm

```bash
# Add required Helm repositories
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

# Create namespace
kubectl create namespace phoenix-system

# Install Phoenix
helm install phoenix ./helm/phoenix \
  --namespace phoenix-system \
  --set global.domain=phoenix.example.com \
  --set newrelic.apiKey.secretName=newrelic-secret
```

### 3. Access the Dashboard

```bash
# Get the dashboard URL
kubectl get ingress -n phoenix-system phoenix-dashboard -o jsonpath='{.spec.rules[0].host}'

# Default credentials
# Username: admin
# Password: changeme (change immediately)
```

## 📖 Documentation

### Getting Started
- [Architecture Overview](docs/architecture.md)
- [User Guide](docs/user-guide.md)
- [Development Guide](docs/DEVELOPMENT.md)
- [Deployment Guide](docs/DEPLOYMENT.md)

### Reference
- [API Reference](docs/api-reference.md)
- [Pipeline Configuration Guide](docs/pipeline-guide.md)
- [Troubleshooting](docs/troubleshooting.md)

### Technical Specifications
- [Product Requirements](docs/PRODUCT_REQUIREMENTS.md)
- [Technical Architecture](docs/TECHNICAL_SPEC_MASTER.md)

## 🛠️ Development

### Prerequisites

```bash
# Install development dependencies
make install-deps

# Setup pre-commit hooks
make setup-hooks
```

### Building from Source

```bash
# Build all components
make build

# Build specific component
make build-api
make build-dashboard
make build-operators
```

### Running Tests

```bash
# Run all tests
make test

# Run specific test suites
make test-unit
make test-integration
make test-e2e
```

### Local Development

```bash
# Start local Kubernetes cluster (using kind)
make cluster-up

# Deploy Phoenix in development mode
make deploy-dev

# Port forward services
make port-forward
```

## 🤝 Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## 📊 Performance

Phoenix is designed to handle:

- 100+ concurrent experiments
- 1000+ nodes per experiment
- 500+ processes per node
- 3.5M+ unique time series

With optimizations, Phoenix typically achieves:

- 50-80% cardinality reduction
- 40-70% cost savings
- <100ms API response time
- <5s pipeline generation time

## 🔒 Security

- JWT-based authentication
- RBAC authorization
- Network policies for pod-to-pod communication
- Secret management via External Secrets Operator
- TLS encryption for all external communication

## 📝 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- OpenTelemetry community for the excellent collector
- New Relic for OTLP endpoint support
- Kubernetes SIG-Apps for operator patterns

## 📞 Support

- Documentation: https://docs.phoenix.io
- Issues: https://github.com/your-org/phoenix-platform/issues
- Slack: #phoenix-support

---

Built with ❤️ by the Phoenix Team