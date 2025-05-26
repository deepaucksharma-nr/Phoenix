# Phoenix Single-VM Troubleshooting Guide

## Quick Diagnostics

Run this first for any issue:
```bash
/opt/phoenix/scripts/health-check.sh
```

## Common Issues & Solutions

### 1. Phoenix UI Not Accessible

#### Symptoms
- Browser shows connection timeout
- SSL certificate errors  
- 502 Bad Gateway

#### Diagnostic Steps
```bash
# Check services
cd /opt/phoenix
sudo docker-compose ps

# Check API logs
sudo docker-compose logs api --tail=50

# Test API health
curl -k https://localhost/health

# Check port binding
sudo netstat -tlnp | grep -E "80|443|6700"
```

#### Solutions
```bash
# Restart services
sudo docker-compose restart

# Regenerate certificates
sudo certbot renew --force-renewal
sudo docker-compose restart api

# Fix permissions
sudo chown -R 1000:1000 /opt/phoenix/data
```

### 2. Agents Not Reporting

#### Symptoms
- Agents showing "offline" in UI
- No heartbeat metrics
- Empty fleet view

#### Diagnostic Steps
```bash
# On Phoenix VM - check connections
sudo tcpdump -i any port 6700 -nn

# Check API logs
sudo docker-compose logs api | grep -i agent

# On Agent - check service
sudo systemctl status phoenix-agent
sudo journalctl -u phoenix-agent -n 50
```

#### Solutions
```bash
# On Agent - restart service
sudo systemctl restart phoenix-agent

# Re-register agent
sudo rm /var/lib/phoenix-agent/agent.id
sudo systemctl restart phoenix-agent

# Check firewall
sudo iptables -L -n | grep 6700
```

### 3. High Memory Usage

#### Symptoms
- VM becomes slow
- Containers restarting
- OOM errors in logs

#### Diagnostic Steps
```bash
# Check memory usage
docker stats --no-stream

# Check for OOM kills
dmesg | grep -i "killed process"

# Database connections
docker-compose exec db psql -U phoenix -c \
  "SELECT count(*) FROM pg_stat_activity;"
```

#### Solutions
```bash
# Increase swap
sudo fallocate -l 4G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile

# Reduce Prometheus retention
# Edit docker-compose.yml
services:
  prometheus:
    command:
      - --storage.tsdb.retention.time=7d

# Restart with memory limits
docker-compose down
docker-compose up -d
```

### 4. Database Issues

#### Symptoms
- "Connection refused" errors
- Slow queries
- Disk full errors

#### Diagnostic Steps
```bash
# Check database status
docker-compose exec db pg_isready

# Check disk space
df -h /opt/phoenix

# Find large tables
docker-compose exec db psql -U phoenix -c "
SELECT tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) 
FROM pg_tables 
WHERE schemaname = 'public' 
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;"
```

#### Solutions
```bash
# Clean up old data
docker-compose exec db psql -U phoenix -c "
DELETE FROM agent_tasks WHERE created_at < NOW() - INTERVAL '7 days';
DELETE FROM metrics_cache WHERE timestamp < NOW() - INTERVAL '30 days';"

# Vacuum database
docker-compose exec db psql -U phoenix -c "VACUUM FULL ANALYZE;"

# Add more disk space
# 1. Add new disk
# 2. Mount to /opt/phoenix/data
# 3. Move data and restart
```

### 5. Metrics Not Showing

#### Symptoms
- Empty graphs in UI
- No data in Prometheus
- Cost calculations failing

#### Diagnostic Steps
```bash
# Check Pushgateway
curl http://localhost:9091/metrics | grep phoenix_

# Check Prometheus targets
curl http://localhost:9090/api/v1/targets

# Query test
curl 'http://localhost:9090/api/v1/query?query=up'
```

#### Solutions
```bash
# Clear Pushgateway
curl -X PUT http://localhost:9091/api/v1/admin/wipe

# Restart metrics pipeline
docker-compose restart prometheus pushgateway

# Force agent metric push
# On each agent:
sudo systemctl restart phoenix-agent
```

### 6. Experiment Stuck

#### Symptoms
- Experiment in "pending" state
- No progress shown
- Agents not picking up tasks

#### Diagnostic Steps
```bash
# Check experiment state
docker-compose exec db psql -U phoenix -c "
SELECT id, phase, error_message 
FROM experiments 
WHERE phase != 'completed';"

# Check task queue
docker-compose exec db psql -U phoenix -c "
SELECT count(*), status 
FROM agent_tasks 
GROUP BY status;"
```

#### Solutions
```bash
# Clear stuck tasks
docker-compose exec db psql -U phoenix -c "
UPDATE agent_tasks 
SET status = 'failed' 
WHERE status = 'pending' 
AND created_at < NOW() - INTERVAL '1 hour';"

# Reset experiment
phoenix experiment reset <experiment-id>

# Force task execution
phoenix task retry --all-failed
```

### 7. Certificate Issues

#### Symptoms
- SSL warnings in browser
- Agent connection failures
- "Certificate expired" errors

#### Diagnostic Steps
```bash
# Check certificate
openssl x509 -in /opt/phoenix/tls/fullchain.pem -text -noout

# Check expiry
openssl x509 -in /opt/phoenix/tls/fullchain.pem -noout -enddate

# Test SSL
openssl s_client -connect phoenix.your-domain.com:443
```

#### Solutions
```bash
# Renew Let's Encrypt
sudo certbot renew --force-renewal

# Generate new self-signed
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /opt/phoenix/tls/privkey.pem \
  -out /opt/phoenix/tls/fullchain.pem \
  -subj "/CN=phoenix.your-domain.com"

# Restart API
docker-compose restart api
```

## Performance Tuning

### Database Optimization
```sql
-- Add missing indexes
CREATE INDEX CONCURRENTLY idx_agent_tasks_status 
  ON agent_tasks(status, created_at);

CREATE INDEX CONCURRENTLY idx_metrics_timestamp 
  ON metrics_cache(timestamp DESC);

-- Update statistics
ANALYZE;
```

### Prometheus Optimization
```yaml
# Edit prometheus.yml
global:
  scrape_interval: 30s  # Increase from 15s
  evaluation_interval: 30s

# Reduce cardinality
metric_relabel_configs:
  - source_labels: [__name__]
    regex: 'go_.*'
    action: drop
```

### API Optimization
```bash
# Increase connection pool
# Edit .env
DATABASE_POOL_SIZE=50
API_WORKERS=4

# Enable response caching
ENABLE_CACHE=true
CACHE_TTL=300
```

## Emergency Procedures

### Complete System Recovery
```bash
#!/bin/bash
# emergency-recovery.sh

# Stop everything
cd /opt/phoenix
docker-compose down

# Clear corrupted data
rm -f data/postgres/postmaster.pid
rm -rf data/prometheus/wal/*

# Restore from backup
./scripts/restore.sh latest

# Start services
docker-compose up -d
```

### Rollback All Changes
```bash
# Stop all experiments
phoenix experiment rollback --all --emergency

# Revert all pipelines
phoenix pipeline rollback --all --force

# Restart agents
phoenix agent restart --all
```

## Log Analysis

### Important Log Locations
```bash
# API logs
docker-compose logs api

# Database logs  
docker-compose logs db

# Agent logs (on each host)
journalctl -u phoenix-agent

# System logs
/var/log/syslog
/var/log/messages
```

### Common Log Patterns
```bash
# Find errors
docker-compose logs | grep -E "ERROR|FATAL|PANIC"

# Agent issues
docker-compose logs api | grep "agent.*failed"

# Database issues
docker-compose logs db | grep -E "FATAL|ERROR"

# Memory issues
dmesg | grep -i memory
```

## Getting Help

1. **Run diagnostics**
   ```bash
   /opt/phoenix/scripts/health-check.sh > diagnostics.txt
   docker-compose logs > logs.txt
   ```

2. **Check documentation**
   - README.md in deployment directory
   - https://docs.phoenix-observability.io

3. **Community support**
   - GitHub Issues: https://github.com/phoenix-observability/phoenix/issues
   - Slack: https://phoenix-community.slack.com

4. **Collect debug bundle**
   ```bash
   # Create support bundle
   tar -czf phoenix-debug-$(date +%s).tar.gz \
     /opt/phoenix/logs \
     /opt/phoenix/.env \
     diagnostics.txt \
     logs.txt
   ```

## Prevention Tips

1. **Monitor regularly**
   - Set up alerts for disk space
   - Monitor memory usage trends
   - Check certificate expiry monthly

2. **Automate maintenance**
   - Daily backups with testing
   - Weekly log rotation
   - Monthly updates

3. **Capacity planning**
   - Track metric growth
   - Plan for 50% headroom
   - Scale before hitting limits