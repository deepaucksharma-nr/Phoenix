
# Phoenix Operations Guide

This document covers day‑to‑day management of the Phoenix Platform. It includes deployment steps for Kubernetes as well as instructions for running the collector on standalone virtual machines.

## Table of Contents

- [Running Collectors on VMs](#running-collectors-on-vms)
- [Systemd Service](#systemd-service)

## Running Collectors on VMs

Collectors can run outside Kubernetes for bare metal or VM environments. Use the CLI to generate a static configuration file:

```bash
# Generate config from a catalog pipeline
phoenix pipeline vm-config process-topk-v1 \
  --exporter-endpoint otel-phoenix.example.com:4317 \
  --output /etc/otelcol/collector.yaml
```

This writes a full OpenTelemetry Collector configuration to `/etc/otelcol/collector.yaml`.

## Systemd Service

Create a unit file `/etc/systemd/system/otelcol.service`:

```ini
[Unit]
Description=Phoenix OTel Collector
After=network.target

[Service]
ExecStart=/usr/local/bin/otelcol --config /etc/otelcol/collector.yaml
Restart=always
User=otel
Group=otel

[Install]
WantedBy=multi-user.target
```

Enable and start the collector:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now otelcol
```

The collector will now run as a background service on the VM.
=======
# Phoenix Platform Operations Guide

This guide describes how to deploy pipelines, run experiments, and analyze results using the Phoenix Platform.

## 1. Deployment Workflow

1. **Bootstrap dependencies**
   ```bash
   make dev-up
   ```
2. **Deploy a pipeline**
   ```bash
   curl -X POST http://localhost:8080/api/v1/pipeline-deployments \
     -H "Content-Type: application/json" \
     -d '{"name":"demo","namespace":"default","template":"process-baseline-v1"}'
   ```
3. **Verify deployment**
   ```bash
   curl http://localhost:8080/api/v1/pipeline-deployments?namespace=default | jq .
   ```

## 2. Experiment Workflow

1. **Create an experiment**
   ```bash
   curl -X POST http://localhost:8080/api/v1/experiments \
     -H "Content-Type: application/json" \
     -d '{"name":"cost-opt","baseline_pipeline":"process-baseline-v1","candidate_pipeline":"process-intelligent-v1","target_namespaces":["default"]}'
   ```
2. **Monitor progress**
   ```bash
   curl http://localhost:8080/api/v1/experiments/<id> | jq .
   ```
3. **Generate configs**
   ```bash
   curl -X POST http://localhost:8082/api/v1/generate \
     -H "Content-Type: application/json" \
     -d '{"experiment_id":"<id>"}'
   ```
4. **Analyze results**
   ```bash
   curl http://localhost:8080/api/v1/experiments/<id>/results | jq .
   ```

## 3. Troubleshooting

- **Check service health**
  ```bash
  curl http://localhost:8080/health
  ```
- **Restart stack**
  ```bash
  make dev-down && make dev-up
  ```

