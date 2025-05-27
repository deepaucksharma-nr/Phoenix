# OpenTelemetry and NRDOT Configuration

## Overview
This directory contains OpenTelemetry and NRDOT collector configuration files for the Phoenix platform.

## Structure
```
./collectors/          # Collector configurations
├── otel/             # Standard OpenTelemetry configs
│   ├── main.yaml     # Main collector config
│   └── observer.yaml # Observer pattern config
└── nrdot/            # NRDOT-specific configs
    ├── baseline.yaml # NRDOT baseline config
    └── optimized.yaml # NRDOT with cardinality reduction

./exporters/          # Exporter configurations
./processors/         # Processor configurations
./receivers/          # Receiver configurations
```

## Collector Types

### OpenTelemetry Collector (Default)
Standard, vendor-neutral OpenTelemetry distribution with Phoenix's custom processors.

### NRDOT Collector
New Relic's optimized OpenTelemetry distribution with:
- Built-in cardinality reduction
- Enhanced performance for New Relic backends
- Native New Relic integration

## Usage
These configurations are used by Phoenix agents based on the `COLLECTOR_TYPE` environment variable.
See individual configuration files for specific details.

## Environment Variables

### Common Variables
- `COLLECTOR_TYPE`: `otel` or `nrdot`
- `EXPERIMENT_ID`: Current experiment ID
- `HOST_ID`: Agent host identifier
- `VARIANT`: `baseline` or `candidate`

### OpenTelemetry Variables
- `OTEL_COLLECTOR_ENDPOINT`: OpenTelemetry collector endpoint
- `PROMETHEUS_ENDPOINT`: Prometheus remote write endpoint

### NRDOT Variables
- `NRDOT_OTLP_ENDPOINT`: NRDOT OTLP endpoint (typically https://otlp.nr-data.net:4317)
- `NEW_RELIC_LICENSE_KEY`: New Relic license key for authentication
- `CARDINALITY_LIMIT`: Maximum metric cardinality (optional)
- `REDUCTION_TARGET`: Target reduction percentage (optional)

## Validation
To validate configurations:
```bash
make validate-configs CONFIG_TYPE=otel
```
