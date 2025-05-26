# Phoenix Capacity Planning Template

## Monthly Capacity Review

**Date**: _____________  
**Reviewer**: _____________

### Current State Metrics

#### Infrastructure
- [ ] VM Type: _____________ (e.g., t3.medium)
- [ ] vCPUs: _____
- [ ] RAM: _____ GB
- [ ] Disk: _____ GB
- [ ] Monthly Cost: $______

#### System Metrics (30-day average)
- [ ] CPU Usage: _____% (threshold: 70%)
- [ ] Memory Usage: _____% (threshold: 80%)
- [ ] Disk Usage: _____% (threshold: 80%)
- [ ] API Latency (p95): _____ ms (threshold: 200ms)

#### Phoenix Metrics
- [ ] Active Agents: _____ (max: 200)
- [ ] Metrics/second: _____K (max: 1M)
- [ ] Database Size: _____ GB
- [ ] Cost Savings: $_____/month
- [ ] Reduction Rate: _____%

### Growth Analysis

#### Historical Growth (Past 3 Months)
| Month | Agents | Metrics/s | DB Size | CPU % | Mem % |
|-------|--------|-----------|---------|-------|-------|
| M-3   |        |           |         |       |       |
| M-2   |        |           |         |       |       |
| M-1   |        |           |         |       |       |
| Now   |        |           |         |       |       |

#### Projected Growth
- [ ] New agents/month: _____
- [ ] Metrics growth rate: _____% per month
- [ ] Database growth: _____ GB per month

### Scaling Triggers Analysis

Check all that apply:

#### Immediate Actions Needed
- [ ] CPU > 85% (Critical)
- [ ] Memory > 90% (Critical)
- [ ] Disk > 90% (Critical)
- [ ] API Latency > 500ms (Critical)
- [ ] Agents > 200 (Critical)

#### Planning Required
- [ ] CPU > 70% (Warning)
- [ ] Memory > 80% (Warning)
- [ ] Disk > 80% (Warning)
- [ ] API Latency > 200ms (Warning)
- [ ] Agents > 150 (Warning)

### Scaling Recommendations

Based on current metrics and growth projections:

#### Next 30 Days
- [ ] No action needed
- [ ] Optimize configuration
  - [ ] Enable aggressive filtering
  - [ ] Increase Prometheus retention
  - [ ] Add swap space
- [ ] Vertical scaling needed
  - [ ] Target VM: _____________
  - [ ] Estimated downtime: _____ minutes

#### Next 90 Days
- [ ] Continue monitoring
- [ ] Plan vertical scaling
- [ ] Plan component separation
  - [ ] PostgreSQL → RDS
  - [ ] Prometheus → Dedicated VM
- [ ] Prepare for horizontal scaling

### Optimization Opportunities

Before scaling, consider:

#### Quick Wins (< 1 hour)
- [ ] Deploy more aggressive metric filters
- [ ] Increase sampling rates
- [ ] Enable API caching
- [ ] Clean up old data

#### Medium Effort (< 1 day)
- [ ] Optimize database indexes
- [ ] Implement metric aggregation
- [ ] Review and optimize Prometheus queries
- [ ] Upgrade to latest Phoenix version

### Action Items

| Priority | Action | Owner | Due Date | Status |
|----------|--------|-------|----------|---------|
| P0       |        |       |          | [ ]     |
| P1       |        |       |          | [ ]     |
| P2       |        |       |          | [ ]     |
| P3       |        |       |          | [ ]     |

### Budget Planning

#### Current Costs
- Infrastructure: $_____/month
- Total with Phoenix savings: $_____/month
- Savings achieved: $_____/month

#### Projected Costs (Next Quarter)
- If no scaling: $_____/month
- With recommended scaling: $_____/month
- Additional budget needed: $_____

### Risk Assessment

Rate each risk (Low/Medium/High):

- [ ] Risk of hitting resource limits: _____
- [ ] Risk of performance degradation: _____
- [ ] Risk of service interruption: _____
- [ ] Risk of data loss: _____

### Notes & Observations

_Use this space for additional context, concerns, or recommendations_

_______________________________________________
_______________________________________________
_______________________________________________
_______________________________________________

### Sign-off

- [ ] Operations Lead: _____________ Date: _____
- [ ] Finance Approval: _____________ Date: _____
- [ ] Technical Review: _____________ Date: _____

---

## Scaling Decision Matrix

Use this matrix to determine the appropriate scaling action:

| Agents | CPU | Memory | Action | Timeline |
|--------|-----|--------|--------|----------|
| < 100  | <50%| <60%   | Monitor only | - |
| < 100  | >70%| -      | Optimize queries | 1 week |
| < 150  | >70%| >80%   | Vertical scale to t3.large | 2 weeks |
| > 150  | >70%| -      | Component separation | 1 month |
| > 150  | -   | >80%   | Component separation | 1 month |
| > 200  | Any | Any    | Horizontal scaling | Immediate |

## Quick Reference

### Scaling Commands
```bash
# Check current capacity
/opt/phoenix/scripts/validate-scaling.sh

# Force metric collection
docker-compose exec api phoenix-cli metrics collect

# Emergency optimization
phoenix pipeline deploy emergency-filter --all --force
```

### Key Contacts
- AWS Support: _______________
- Phoenix Support: _______________
- On-call Engineer: _______________

Remember: **Optimize first, scale second!**