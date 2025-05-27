# Creating Your First Experiment

This guide walks you through creating and running your first cost optimization experiment with Phoenix.

## Prerequisites

Before starting, ensure you have:
- Phoenix Platform running ([Quick Start Guide](../../QUICKSTART.md))
- Access to the Dashboard (http://localhost:3000)
- Phoenix CLI installed and configured
- At least one Phoenix Agent connected

## Step 1: Understand Your Current Metrics

First, check your current metrics cardinality and costs:

```bash
# Check fleet status
phoenix-cli fleet status

# View current metrics rate
phoenix-cli metrics summary
```

Expected output:
```
Fleet Status:
- Active Agents: 10
- Total Metrics/sec: 500,000
- Estimated Monthly Cost: $3,750

Top Cardinality Sources:
1. app_http_requests_total: 125,000 series (25%)
2. container_cpu_usage: 75,000 series (15%)
3. custom_business_metrics: 50,000 series (10%)
```

## Step 2: Choose a Pipeline Strategy

Phoenix offers three optimization strategies:

### Adaptive Filter (Recommended for first experiment)
- Uses ML to identify important metrics
- Maintains critical business metrics
- Typically achieves 60-80% reduction

### TopK
- Keeps only the top K most active metrics
- Simple and predictable
- Best for high-volume, low-value metrics

### Hybrid
- Combines multiple strategies
- Maximum reduction potential
- Requires more tuning

For this guide, we'll use the Adaptive Filter.

### Collector Selection

You can run experiments with either:
- **OpenTelemetry Collector** (default) - Vendor-neutral, community standard
- **NRDOT Collector** - New Relic's optimized distribution with enhanced cardinality reduction

For New Relic users, NRDOT provides additional optimization benefits.

## Step 3: Create the Experiment

### Using the Dashboard

1. Navigate to **Experiments** → **New Experiment**
2. Fill in the configuration:
   - **Name**: "First Cardinality Reduction"
   - **Description**: "Test adaptive filter on application metrics"
   - **Baseline Pipeline**: "default" (current configuration)
   - **Candidate Pipeline**: "adaptive-filter-v2"
   - **Traffic Split**: 20% (start conservative)
   - **Duration**: 2 hours

3. Set success criteria:
   - **Minimum Cardinality Reduction**: 60%
   - **Maximum Signal Loss**: 1%
   - **Maximum Latency Increase**: 10ms

4. Click **Create Experiment**

### Using the CLI

```bash
phoenix-cli experiment create \
  --name "First Cardinality Reduction" \
  --description "Test adaptive filter on application metrics" \
  --baseline-pipeline default \
  --candidate-pipeline adaptive-filter-v2 \
  --traffic-split 20 \
  --duration 2h \
  --min-reduction 0.6 \
  --max-signal-loss 0.01 \
  --max-latency-ms 10
```

For NRDOT users:
```bash
phoenix-cli experiment create \
  --name "NRDOT Cardinality Reduction" \
  --description "Test NRDOT adaptive filter" \
  --baseline-pipeline default \
  --candidate-pipeline nrdot-adaptive-filter \
  --traffic-split 20 \
  --duration 2h \
  --min-reduction 0.6 \
  --collector-type nrdot \
  --config '{"nr_license_key": "${NEW_RELIC_LICENSE_KEY}"}'
```

Output:
```
Experiment created successfully!
ID: exp-a1b2c3d4
Status: created
Next: Start the experiment with 'phoenix-cli experiment start exp-a1b2c3d4'
```

## Step 4: Start the Experiment

### Using the Dashboard

1. Go to **Experiments** → **exp-a1b2c3d4**
2. Review the configuration
3. Click **Start Experiment**
4. Confirm agent selection (2 agents for 20% split)

### Using the CLI

```bash
phoenix-cli experiment start exp-a1b2c3d4
```

Output:
```
Starting experiment exp-a1b2c3d4...
✓ Deploying baseline to 8 agents
✓ Deploying candidate to 2 agents
✓ Experiment started successfully

Monitor progress:
- Dashboard: http://localhost:3000/experiments/exp-a1b2c3d4
- CLI: phoenix-cli experiment status exp-a1b2c3d4
```

## Step 5: Monitor Progress

### Real-time Monitoring (Dashboard)

The experiment dashboard shows:
- **Progress Bar**: Time elapsed vs. total duration
- **Live Metrics**: 
  - Baseline vs. Candidate metrics/second
  - Current reduction percentage
  - Error rates
- **Cost Savings**: Projected monthly savings
- **Agent Status**: Health of all agents

### CLI Monitoring

```bash
# Get current status
phoenix-cli experiment status exp-a1b2c3d4

# Watch metrics in real-time
phoenix-cli experiment metrics exp-a1b2c3d4 --watch

# View detailed analysis
phoenix-cli experiment analyze exp-a1b2c3d4
```

Example output:
```
Experiment: First Cardinality Reduction
Status: running (45% complete)
Duration: 54m / 2h

Metrics Summary:
├─ Baseline:  500,000 metrics/sec (8 agents)
├─ Candidate: 150,000 metrics/sec (2 agents)
└─ Reduction: 70.0% ✓

Quality Metrics:
├─ Signal Preservation: 99.8% ✓
├─ Error Rate: 0.01% ✓
└─ P99 Latency: +3ms ✓

Projected Savings: $2,625/month
```

## Step 6: Analyze Results

After the experiment completes:

### Success Indicators
- ✅ Cardinality reduced by target amount
- ✅ All quality metrics within thresholds
- ✅ No increase in error rates
- ✅ Agents remained healthy

### Review Detailed Metrics

```bash
# Get final report
phoenix-cli experiment report exp-a1b2c3d4 --format detailed
```

The report includes:
- Time-series graphs of all metrics
- Statistical analysis
- Specific metrics that were filtered
- Cost savings breakdown

## Step 7: Make a Decision

Based on results, you have three options:

### 1. Promote to Production (Success)

If the experiment met all success criteria:

```bash
# Gradually roll out to all agents
phoenix-cli experiment promote exp-a1b2c3d4 \
  --strategy gradual \
  --rate 10 \
  --interval 1h
```

This will:
- Increase deployment by 10% every hour
- Monitor for issues during rollout
- Automatically pause if problems detected

### 2. Rollback (Issues Detected)

If problems were found:

```bash
phoenix-cli experiment rollback exp-a1b2c3d4 \
  --reason "Higher than expected signal loss"
```

### 3. Iterate (Partial Success)

If results are promising but need tuning:

```bash
# Create new experiment with adjusted parameters
phoenix-cli experiment clone exp-a1b2c3d4 \
  --name "Cardinality Reduction v2" \
  --param importance_threshold=0.8 \
  --param evaluation_interval=5m
```

## Step 8: Monitor Production

After promotion, monitor the full deployment:

### Set Up Alerts

```bash
# Create alert for cardinality spikes
phoenix-cli alert create \
  --name "Cardinality Spike Detection" \
  --condition "cardinality_reduction < 0.5" \
  --severity warning \
  --channel slack
```

### Track Long-term Savings

View accumulated savings in the dashboard:
- **Cost Flow**: Real-time savings visualization
- **Monthly Reports**: Detailed cost breakdowns
- **ROI Tracking**: Platform value metrics

## Common Issues and Solutions

### High Signal Loss
**Problem**: Important metrics being filtered
**Solution**: Adjust importance threshold higher
```bash
--param importance_threshold=0.85
```

### Insufficient Reduction
**Problem**: Not achieving target reduction
**Solution**: Try TopK strategy for specific namespaces
```bash
--candidate-pipeline topk-sampler
--param k=1000
--param namespace_filter="container_*"
```

### Agent Failures
**Problem**: Agents disconnecting during experiment
**Solution**: Check agent logs and resources
```bash
phoenix-cli agent logs agent-hostname-001
phoenix-cli agent health --verbose
```

## Best Practices

1. **Start Small**: Use 10-20% traffic split initially
2. **Short Duration**: 1-2 hours for first experiments
3. **Monitor Actively**: Watch the first 30 minutes closely
4. **Document Results**: Keep notes on what worked
5. **Iterate Quickly**: Run multiple experiments to find optimal settings

## Next Steps

Congratulations on running your first experiment! Next:

1. [Explore different pipeline strategies](../user-guide/pipelines.md)
2. [Set up automated monitoring](../user-guide/monitoring.md)
3. [Learn advanced experiment techniques](../tutorials/reduce-cardinality.md)
4. [Build custom pipelines](../tutorials/custom-pipelines.md)

## Getting Help

- **Dashboard**: Built-in help tooltips
- **CLI**: `phoenix-cli experiment --help`
- **Community**: [Discord](https://discord.gg/phoenix)
- **Support**: support@phoenix.io