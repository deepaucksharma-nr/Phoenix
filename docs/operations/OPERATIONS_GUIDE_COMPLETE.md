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

