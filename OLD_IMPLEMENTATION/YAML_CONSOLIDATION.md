# Phoenix-vNext YAML Consolidation Guide

## Overview

All YAML/YML configuration files have been consolidated to eliminate redundancy and create a clear, maintainable structure.

## Consolidated Structure

### 1. OpenTelemetry Configurations

**Canonical Location:** `configs/otel/`

```
configs/otel/
├── collectors/
│   ├── main.yaml          # Main collector configuration
│   └── observer.yaml      # Observer collector configuration
├── processors/            # Shared processor configurations
└── exporters/
    └── newrelic-enhanced.yaml  # New Relic exporter config
```

**Production Variant:** `configs/production/otel_collector_main_prod.yaml`
- Includes TLS, authentication, and production-specific settings

### 2. Prometheus Configuration

**Canonical Location:** `configs/monitoring/prometheus/`

```
configs/monitoring/prometheus/
├── prometheus.yaml        # Main Prometheus configuration
└── rules/
    └── phoenix_rules.yml  # Consolidated recording rules and alerts
```

**What was consolidated:**
- Merged 5+ rule files into single comprehensive `phoenix_rules.yml`
- Includes operational alerts, performance metrics, ML features, and SLIs

### 3. Control Configuration

**Canonical Location:** `configs/control/`

```
configs/control/
└── optimization_mode.yaml  # Dynamic control state file
```

### 4. Docker Compose Files

**Main Files:**
```
docker-compose.yaml         # Base configuration for all services
docker-compose.override.yml # Development overrides (auto-loaded)
docker-compose.prod.yml     # Production overrides (manual load)
```

**Usage:**
```bash
# Development (auto-loads override)
docker-compose up -d

# Production
docker-compose -f docker-compose.yaml -f docker-compose.prod.yml up -d
```

### 5. Service-Specific Configurations

**Retained Service Configs:**
- `services/validator/config/config.yaml` - Benchmark validation settings
- `packages/contracts/openapi/control-api.yaml` - API specifications

### 6. CI/CD Configurations

**GitHub Actions:** `.github/workflows/`
- `ci.yml` - Continuous integration
- `security.yml` - Security scanning

## Removed Redundancies

### Archived Files
All redundant files moved to `archive/yaml-consolidation/`:

1. **Control Configs:**
   - `tools/configs/control/*` → Archived

2. **Prometheus Rules:**
   - `phoenix_advanced_rules.yml` → Merged into main rules
   - `tools/configs/monitoring/prometheus/rules/*` → Archived

3. **Collector Configs:**
   - `services/collector/configs/*` → Archived
   - `services/control-plane/observer/config/*` → Archived

4. **Docker Compose:**
   - `docker-compose.dev.yml` → Merged into override
   - `infrastructure/docker/*` → Archived

## Benefits

### 1. Single Source of Truth
- Each configuration type has one canonical location
- No confusion about which file to edit

### 2. Simplified Maintenance
- Updates only needed in one place
- Clear inheritance hierarchy (base → override → prod)

### 3. Better Performance
- Consolidated Prometheus rules = faster evaluation
- Single config load = faster startup

### 4. Clearer Dependencies
- Explicit file references in docker-compose
- No circular dependencies

## Migration Guide

### For Existing Deployments

1. **Update File References:**
   ```bash
   # Old
   ./tools/configs/control/optimization_mode.yaml
   
   # New
   ./configs/control/optimization_mode.yaml
   ```

2. **Update Docker Commands:**
   ```bash
   # Old (multiple compose files)
   docker-compose -f docker-compose.yaml -f docker-compose.dev.yml up
   
   # New (automatic override)
   docker-compose up
   ```

3. **Update CI/CD Pipelines:**
   - Point to new config locations
   - Remove references to archived files

### For New Deployments

Simply use the canonical locations:
- OTel configs: `configs/otel/collectors/`
- Prometheus: `configs/monitoring/prometheus/`
- Control: `configs/control/`

## Best Practices

### 1. Environment-Specific Overrides
- Base config: `docker-compose.yaml`
- Dev overrides: `docker-compose.override.yml`
- Prod overrides: `docker-compose.prod.yml`

### 2. Configuration Hierarchy
```
Base Configuration (checked into git)
    ↓
Environment Variables (.env file)
    ↓
Runtime Overrides (command line)
```

### 3. Adding New Services
1. Add to `docker-compose.yaml`
2. Add dev settings to `docker-compose.override.yml`
3. Add prod settings to `docker-compose.prod.yml`

## Validation

Run these commands to validate the consolidation:

```bash
# Check YAML syntax
find configs -name "*.yml" -o -name "*.yaml" | xargs -I {} yamllint {}

# Validate docker-compose
docker-compose config

# Check for remaining duplicates
find . -name "*.yml" -o -name "*.yaml" | grep -v archive | sort | uniq -d
```

## Archive Structure

Archived files are organized by type:
```
archive/yaml-consolidation/
├── control/              # Old control configs
├── prometheus/           # Old Prometheus rules
├── collectors/           # Old collector configs
├── docker/              # Old docker-compose variants
└── tools-configs/       # Old tools directory configs
```

These can be safely deleted after validation.