# Phoenix Helm Chart Configuration Clarification

## Overview
This document clarifies the Phoenix Helm chart configuration, addressing confusion around the API Gateway vs API Service deployment identified in the functional review.

## Current Architecture

### Service Components

1. **API Service** (`api`)
   - The core Phoenix API service (gRPC + HTTP via grpc-gateway)
   - Handles all experiment management, pipeline operations
   - Serves both gRPC (port 5050) and HTTP REST (port 8080)
   - This is the actual Phoenix backend service

2. **API Gateway** (`apiGateway`)
   - External Kong gateway (optional)
   - Provides additional features: rate limiting, API key management, advanced routing
   - Currently configured but not strictly required for MVP
   - Adds complexity for initial deployments

3. **Experiment Controller** (`experimentController`)
   - Kubernetes controller for experiment orchestration
   - Watches experiment state and coordinates deployments
   - Separate from the API service (not combined)

## Deployment Options

### Option 1: Direct API Service (Recommended for MVP)
```yaml
# Simplified deployment - API service directly exposed
api:
  enabled: true
  service:
    type: LoadBalancer  # Or use Ingress
    
apiGateway:
  enabled: false  # Disable Kong for simplicity
  
# Add ingress for API service
apiIngress:
  enabled: true
  className: nginx
  hosts:
    - host: api.phoenix.example.com
      paths:
        - path: /
          pathType: Prefix
          backend:
            service:
              name: phoenix-api
              port:
                number: 8080
```

### Option 2: With Kong Gateway (Production)
```yaml
# Full deployment with Kong
api:
  enabled: true
  service:
    type: ClusterIP  # Internal only
    
apiGateway:
  enabled: true
  upstream:
    - name: phoenix-api
      host: phoenix-api
      port: 8080
      
# Kong handles external traffic
```

## Recommended Changes

### 1. Update values.yaml
```yaml
# Add explicit API ingress configuration
api:
  # ... existing config ...
  
  ingress:
    enabled: false  # Set to true if not using Kong
    className: nginx
    annotations:
      cert-manager.io/cluster-issuer: letsencrypt-prod
      nginx.ingress.kubernetes.io/backend-protocol: "GRPC"
      nginx.ingress.kubernetes.io/grpc-backend: "true"
    hosts:
      - host: api.phoenix.example.com
        paths:
          - path: /
            pathType: Prefix
    tls:
      - secretName: api-tls
        hosts:
          - api.phoenix.example.com

# Make Kong optional with clear flag
apiGateway:
  enabled: false  # Disabled by default for simplicity
  # ... rest of Kong config ...
```

### 2. Add Deployment Decision Helper
```yaml
# deployment-modes.yaml
deploymentMode: "simple"  # Options: simple, production

# Simple mode settings
simple:
  api:
    service:
      type: LoadBalancer
  apiGateway:
    enabled: false
  monitoring:
    minimal: true

# Production mode settings  
production:
  api:
    service:
      type: ClusterIP
  apiGateway:
    enabled: true
  monitoring:
    full: true
```

### 3. Update Templates

#### api-deployment.yaml
```yaml
{{- if .Values.api.enabled }}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "phoenix.fullname" . }}-api
  labels:
    app.kubernetes.io/component: api
    {{- include "phoenix.labels" . | nindent 4 }}
spec:
  # ... deployment spec
{{- end }}
```

#### api-ingress.yaml (New)
```yaml
{{- if and .Values.api.enabled .Values.api.ingress.enabled }}
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{ include "phoenix.fullname" . }}-api
  labels:
    {{- include "phoenix.labels" . | nindent 4 }}
  annotations:
    {{- toYaml .Values.api.ingress.annotations | nindent 4 }}
spec:
  {{- if .Values.api.ingress.className }}
  ingressClassName: {{ .Values.api.ingress.className }}
  {{- end }}
  {{- if .Values.api.ingress.tls }}
  tls:
    {{- toYaml .Values.api.ingress.tls | nindent 4 }}
  {{- end }}
  rules:
    {{- range .Values.api.ingress.hosts }}
    - host: {{ .host }}
      http:
        paths:
          {{- range .paths }}
          - path: {{ .path }}
            pathType: {{ .pathType }}
            backend:
              service:
                name: {{ include "phoenix.fullname" $ }}-api
                port:
                  number: {{ $.Values.api.service.httpPort }}
          {{- end }}
    {{- end }}
{{- end }}
```

## Installation Guide Updates

### Quick Start (Development)
```bash
# Install Phoenix with simple mode
helm install phoenix ./helm/phoenix \
  --set deploymentMode=simple \
  --set api.service.type=NodePort \
  --set dashboard.ingress.enabled=false
  
# Port forward for local access
kubectl port-forward svc/phoenix-api 8080:8080
kubectl port-forward svc/phoenix-dashboard 3000:80
```

### Production Installation
```bash
# Install Phoenix with Kong gateway
helm install phoenix ./helm/phoenix \
  --set deploymentMode=production \
  --set apiGateway.enabled=true \
  --set global.domain=phoenix.company.com \
  --set-file apiGateway.license=/path/to/kong-license
```

## Architecture Diagram Update

```
┌─────────────┐     ┌─────────────┐
│   Browser   │────▶│  Dashboard  │
└─────────────┘     └─────────────┘
                           │
                           ▼
┌─────────────┐     ┌─────────────┐
│   CLI/API   │────▶│ API Service │
│   Clients   │     │  (HTTP/gRPC)│
└─────────────┘     └─────────────┘
                           │
        ┌──────────────────┼──────────────────┐
        ▼                  ▼                  ▼
┌───────────────┐ ┌───────────────┐ ┌───────────────┐
│  Experiment   │ │    Config     │ │   Pipeline    │
│  Controller   │ │   Generator   │ │   Operator    │
└───────────────┘ └───────────────┘ └───────────────┘
```

With Kong (Optional):
```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Clients   │────▶│Kong Gateway │────▶│ API Service │
└─────────────┘     └─────────────┘     └─────────────┘
```

## Benefits of This Approach

1. **Flexibility**: Users can start simple and add Kong later
2. **Clarity**: Clear separation between required and optional components  
3. **Reduced Complexity**: MVP deployments don't need Kong
4. **Future-Proof**: Easy to enable Kong when needed

## Migration Path

For users who start with simple mode and want to add Kong later:

1. Update values to enable Kong
2. Update ingress to point to Kong instead of API
3. Configure Kong upstream to API service
4. Apply Helm upgrade

```bash
# Upgrade from simple to production
helm upgrade phoenix ./helm/phoenix \
  --set apiGateway.enabled=true \
  --set api.ingress.enabled=false \
  --reuse-values
```

## Conclusion

This clarification:
- Removes ambiguity about API service vs gateway roles
- Provides clear deployment options
- Simplifies initial setup while maintaining flexibility
- Documents the architecture clearly for users

The recommended approach is to start with the simple deployment (no Kong) for most users, adding the API gateway only when advanced features are needed.