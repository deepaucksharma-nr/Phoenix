# Phoenix CLI Workflow Examples

This directory contains example scripts demonstrating various Phoenix CLI workflows and integration patterns.

## Available Examples

### 1. experiment-workflow.sh
Complete experiment lifecycle management from creation to promotion.
- Authentication and setup
- Experiment creation with traffic splitting
- Real-time monitoring
- Metrics analysis
- Automated promotion based on criteria
- Configuration export

### 2. pipeline-deployment-workflow.sh
Direct pipeline deployment without experiments.
- Deploy specific pipeline templates
- Monitor deployment progress
- Update configurations
- Rollback capabilities
- Export deployment configurations

### 3. batch-operations.sh
Bulk operations and automation patterns.
- Create multiple experiments programmatically
- Parallel operations
- Automated decision making based on metrics
- Bulk configuration export
- Summary report generation

### 4. cicd-integration.sh
Integration with CI/CD pipelines (Jenkins, GitLab CI, GitHub Actions).
- Environment-based experiment strategies
- Quality gate implementation
- Automated testing during experiments
- Build-specific experiment tracking
- Automated promotion/rollback decisions

### 5. monitoring-dashboard.sh
Real-time monitoring dashboard in the terminal.
- Live experiment status updates
- Metrics visualization with color coding
- Progress bars and formatted output
- System-wide metrics aggregation
- Recent events display

## Running the Examples

1. Make scripts executable:
```bash
chmod +x *.sh
```

2. Set environment variables (optional):
```bash
export PHOENIX_API_URL="http://localhost:8080"
export PHOENIX_API_TOKEN="your-token"  # For CI/CD
```

3. Run any example:
```bash
./experiment-workflow.sh
```

## Common Patterns

### Authentication
All scripts check authentication status and prompt for login if needed:
```bash
if ! phoenix auth status >/dev/null 2>&1; then
    phoenix auth login
fi
```

### Error Handling
Scripts use `set -e` to exit on errors and provide meaningful error messages.

### Output Parsing
JSON output is parsed using `jq` for reliable data extraction:
```bash
EXPERIMENT_ID=$(phoenix experiment create ... --output json | jq -r '.id')
```

### Parallel Operations
Background jobs with `wait` for parallel execution:
```bash
for exp_id in "${EXPERIMENTS[@]}"; do
    phoenix experiment start "$exp_id" &
done
wait
```

## Customization

These scripts are templates that can be customized for your specific use cases:

- Modify traffic split ratios
- Adjust quality gate thresholds
- Add custom notification integrations
- Integrate with your monitoring systems
- Extend with custom metrics analysis

## Requirements

- Phoenix CLI installed and in PATH
- `jq` for JSON parsing
- `bc` for floating-point calculations
- Basic Unix utilities (grep, sed, awk)
- Terminal with color support (for monitoring dashboard)

## Integration Tips

### Jenkins
```groovy
sh '''
    export PHOENIX_API_TOKEN="${PHOENIX_TOKEN}"
    ./cicd-integration.sh
'''
```

### GitLab CI
```yaml
phoenix-experiment:
  script:
    - export PHOENIX_API_TOKEN="$PHOENIX_TOKEN"
    - ./cicd-integration.sh
  artifacts:
    paths:
      - phoenix-reports-*/
```

### GitHub Actions
```yaml
- name: Run Phoenix Experiment
  env:
    PHOENIX_API_TOKEN: ${{ secrets.PHOENIX_TOKEN }}
  run: ./cicd-integration.sh
```

## Best Practices

1. **Always validate** pipeline configurations before deployment
2. **Use meaningful names** for experiments and deployments
3. **Set appropriate timeouts** for monitoring loops
4. **Implement proper cleanup** for failed experiments
5. **Log all automated decisions** with clear reasons
6. **Use version control** for pipeline configurations
7. **Monitor quality gates** continuously during experiments
8. **Export configurations** for reproducibility

## Troubleshooting

- **Authentication failures**: Check API URL and credentials
- **JSON parsing errors**: Ensure `jq` is installed
- **Permission denied**: Make scripts executable with `chmod +x`
- **API timeouts**: Increase timeout values or check connectivity
- **Missing commands**: Install required dependencies (jq, bc)