# monitoring Configuration

## Overview
This directory contains monitoring configuration files for the Phoenix platform.

## Structure
./grafana
./grafana/dashboards
./grafana/dashboards_provider.yaml
./grafana/grafana-datasource.yaml
./grafana/provisioning
./prometheus
./prometheus/alerts
./prometheus/prometheus.yaml
./prometheus/rules

## Usage
These configurations are used by various services in the Phoenix platform.
See individual configuration files for specific details.

## Environment Variables
Configuration files may reference environment variables for sensitive values.
Ensure all required variables are set before deployment.

## Validation
To validate configurations:
```bash
make validate-configs CONFIG_TYPE=monitoring
```
