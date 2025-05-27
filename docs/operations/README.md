# Phoenix Platform Operations Guide

This directory contains comprehensive operational documentation for the Phoenix Platform.

## Quick Navigation

### 🚀 Getting Started
- [Configuration Reference](configuration.md) - All configuration options
- [Docker Compose Setup](docker-compose.md) - Container deployment guide
- [Complete Operations Guide](OPERATIONS_GUIDE_COMPLETE.md) - Comprehensive operations manual

### 🔌 Collector Integration
- [NRDOT Integration Guide](nrdot-integration.md) - New Relic collector setup
- [NRDOT Troubleshooting](nrdot-troubleshooting.md) - Common issues and solutions

### 🏗️ Deployment
- [Single-VM Deployment](../../deployments/single-vm/README.md) - Production deployment guide
- [Environment Variables](../../.env.example) - Configuration template

## Operations Overview

### System Requirements

#### Minimum Requirements (Development)
- 2 vCPU
- 4GB RAM
- 20GB disk space
- Docker & Docker Compose

#### Recommended Requirements (Production)
- 4 vCPU
- 16GB RAM
- 100GB SSD storage
- Docker & Docker Compose
- PostgreSQL 15+

### Key Operational Tasks

#### 1. Initial Setup
```bash
# Clone repository
git clone https://github.com/phoenix/platform.git
cd platform

# Setup environment
cp .env.example .env
# Edit .env with your configuration

# Start services
docker-compose up -d
```

#### 2. Health Monitoring
```bash
# Check service health
docker-compose ps

# View logs
docker-compose logs -f phoenix-api

# Check metrics
curl http://localhost:9090/metrics
```

#### 3. Backup & Recovery
```bash
# Backup database
pg_dump -h localhost -U phoenix -d phoenix > backup.sql

# Restore database
psql -h localhost -U phoenix -d phoenix < backup.sql
```

#### 4. Scaling Operations
- Horizontal scaling: Add more Phoenix API instances
- Agent scaling: Deploy agents to more hosts
- Database scaling: Use external PostgreSQL with replicas

### Monitoring Stack

The Phoenix Platform includes integrated monitoring:

- **Prometheus**: Metrics collection (port 9090)
- **Grafana**: Visualization (port 3001)
- **Pushgateway**: Metric aggregation (port 9091)

### Security Considerations

1. **Network Security**
   - Use TLS for all external communications
   - Implement firewall rules for service ports
   - Use Docker networks for service isolation

2. **Authentication & Authorization**
   - JWT tokens for user authentication
   - X-Agent-Host-ID for agent authentication
   - Role-based access control (RBAC)

3. **Data Protection**
   - Encrypt data at rest (PostgreSQL)
   - Use environment variables for secrets
   - Regular security updates

### Troubleshooting Quick Reference

| Issue | Solution |
|-------|----------|
| API not responding | Check `docker-compose logs phoenix-api` |
| Agent can't connect | Verify `PHOENIX_API_URL` and network connectivity |
| High memory usage | Check collector configuration and cardinality |
| Database errors | Verify PostgreSQL connection and permissions |

### Performance Tuning

#### API Performance
```yaml
# Increase connection pool
DATABASE_MAX_CONNECTIONS: 100
DATABASE_MAX_IDLE: 10

# Enable caching
REDIS_URL: redis://redis:6379
```

#### Agent Performance
```yaml
# Reduce polling interval
POLL_INTERVAL: 30s

# Limit concurrent collectors
MAX_COLLECTORS: 50
```

### Maintenance Windows

Recommended maintenance schedule:
- **Daily**: Log rotation, metric cleanup
- **Weekly**: Database vacuum, backup verification
- **Monthly**: Security updates, performance review
- **Quarterly**: Major version upgrades

## Related Documentation

- [Architecture Overview](../../ARCHITECTURE.md)
- [Development Guide](../../DEVELOPMENT_GUIDE.md)
- [API Documentation](../api/README.md)
- [Deployment Guide](../../deployments/single-vm/README.md)

## Support

For operational support:
1. Check [Troubleshooting Guide](nrdot-troubleshooting.md)
2. Review [Operations Guide](OPERATIONS_GUIDE_COMPLETE.md)
3. Search [GitHub Issues](https://github.com/phoenix/platform/issues)
4. Join [Discord Community](https://discord.gg/phoenix)