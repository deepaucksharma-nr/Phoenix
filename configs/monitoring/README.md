# Phoenix Monitoring Configuration

## Overview
This directory contains monitoring configuration files for the Phoenix platform's cost optimization system.

## Phoenix Platform Monitoring
The Phoenix platform implements intelligent metric filtering to reduce observability costs by 70% while maintaining critical visibility.

## Structure
```
monitoring/
├── grafana/
│   ├── dashboards/         # Cost reduction dashboards
│   └── provisioning/      # Grafana configuration
├── prometheus/
│   ├── prometheus.yml     # Scraping configuration
│   └── rules/            # Alerting rules
└── README.md
```

## Key Metrics Tracked
- Metric cardinality reduction percentage
- Cost savings (monthly/yearly)
- Agent health and status
- Experiment progress
- Pipeline performance

## Usage
These configurations support:
- Real-time cost monitoring dashboards
- Agent fleet status visualization
- Experiment A/B testing metrics
- Cardinality reduction tracking

## Environment Variables
```bash
PROMETHEUS_URL=http://localhost:9090
PUSHGATEWAY_URL=http://localhost:9091
GRAFANA_ADMIN_PASSWORD=secure-password
```

## Dashboards
- **Phoenix Overview**: Cost savings and agent status
- **Experiment Tracking**: A/B test results
- **Cardinality Analysis**: Metric reduction details
- **Fleet Management**: Agent health monitoring

## Integration
Configurations work with:
- Phoenix API (port 8080)
- WebSocket real-time updates
- Agent task polling system
- PostgreSQL metrics storage
