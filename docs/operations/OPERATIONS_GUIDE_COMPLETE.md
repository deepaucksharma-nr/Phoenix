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
