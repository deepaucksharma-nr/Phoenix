#!/bin/bash

# Fix Docker Compose configuration issues

set -e

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
PROJECT_ROOT=$(cd "$SCRIPT_DIR/.." && pwd)

echo "Fixing Docker Compose configuration..."

# Create a temporary docker-compose file with fixed paths
cat > "$PROJECT_ROOT/docker-compose-fixed.yml" << 'EOF'
version: '3.9'

services:
  # PostgreSQL Database
  postgres:
    image: postgres:15-alpine
    container_name: phoenix-postgres
    restart: unless-stopped
    environment:
      POSTGRES_USER: phoenix
      POSTGRES_PASSWORD: phoenix
      POSTGRES_DB: phoenix_db
      POSTGRES_INITDB_ARGS: "-E UTF8"
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U phoenix"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - phoenix-network

  # Redis Cache
  redis:
    image: redis:7-alpine
    container_name: phoenix-redis
    restart: unless-stopped
    command: redis-server --appendonly yes --requirepass phoenix
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "--pass", "phoenix", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - phoenix-network

  # NATS Message Queue
  nats:
    image: nats:2.10-alpine
    container_name: phoenix-nats
    restart: unless-stopped
    command: ["-js", "-m", "8222"]
    ports:
      - "4222:4222"  # Client connections
      - "8222:8222"  # Monitoring
      - "6222:6222"  # Cluster
    volumes:
      - nats_data:/data
    healthcheck:
      test: ["CMD", "nc", "-z", "localhost", "4222"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - phoenix-network

  # Prometheus
  prometheus:
    image: prom/prometheus:v2.47.0
    container_name: phoenix-prometheus
    restart: unless-stopped
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/usr/share/prometheus/console_libraries'
      - '--web.console.templates=/usr/share/prometheus/consoles'
      - '--web.enable-lifecycle'
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus/prometheus.yaml:/etc/prometheus/prometheus.yml:ro
      - ./monitoring/prometheus/rules/:/etc/prometheus/rules/:ro
      - prometheus_data:/prometheus
    networks:
      - phoenix-network

  # Grafana
  grafana:
    image: grafana/grafana:10.2.0
    container_name: phoenix-grafana
    restart: unless-stopped
    environment:
      GF_SECURITY_ADMIN_USER: admin
      GF_SECURITY_ADMIN_PASSWORD: phoenix
      GF_USERS_ALLOW_SIGN_UP: false
      GF_SERVER_ROOT_URL: http://localhost:3000
      GF_SMTP_ENABLED: false
      GF_LOG_LEVEL: warn
    ports:
      - "3000:3000"
    volumes:
      - ./monitoring/grafana/dashboards:/var/lib/grafana/dashboards:ro
      - grafana_data:/var/lib/grafana
    depends_on:
      - prometheus
    networks:
      - phoenix-network

  # Jaeger Tracing
  jaeger:
    image: jaegertracing/all-in-one:1.50
    container_name: phoenix-jaeger
    restart: unless-stopped
    environment:
      COLLECTOR_ZIPKIN_HOST_PORT: :9411
      COLLECTOR_OTLP_ENABLED: true
    ports:
      - "5775:5775/udp"
      - "6831:6831/udp"
      - "6832:6832/udp"
      - "5778:5778"
      - "16686:16686"  # UI
      - "14268:14268"
      - "14250:14250"
      - "9411:9411"
      - "4317:4317"    # OTLP gRPC
      - "4318:4318"    # OTLP HTTP
    networks:
      - phoenix-network

  # OpenTelemetry Collector (simplified config)
  otel-collector:
    image: otel/opentelemetry-collector:0.88.0
    container_name: phoenix-otel-collector
    restart: unless-stopped
    command: ["--config=/etc/otel-collector-config.yaml"]
    environment:
      - OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger:4317
    ports:
      - "1888:1888"   # pprof extension
      - "8888:8888"   # Prometheus metrics exposed by the collector
      - "8889:8889"   # Prometheus exporter metrics
      - "13133:13133" # health_check extension
      - "55679:55679" # zpages extension
    depends_on:
      - jaeger
      - prometheus
    networks:
      - phoenix-network

volumes:
  postgres_data:
  redis_data:
  nats_data:
  prometheus_data:
  grafana_data:

networks:
  phoenix-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.28.0.0/16
EOF

echo "Fixed docker-compose file created at docker-compose-fixed.yml"
echo "Use: docker-compose -f docker-compose-fixed.yml up -d"