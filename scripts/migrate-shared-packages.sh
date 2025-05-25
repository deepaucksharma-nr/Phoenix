#!/bin/bash
# migrate-shared-packages.sh - Extract and migrate shared code to pkg/ directory
# Usage: ./migrate-shared-packages.sh

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Phoenix Shared Package Migration ===${NC}"
echo ""

# Base paths
OLD_PKG="OLD_IMPLEMENTATION/phoenix-platform/pkg"
NEW_PKG="pkg"

# Create pkg directory structure
echo -e "${YELLOW}Creating shared package structure...${NC}"
mkdir -p "$NEW_PKG"/{auth,database,messaging,telemetry,http,grpc,k8s,testing,errors,utils,config}
mkdir -p "$NEW_PKG"/auth/{jwt,oauth,rbac}
mkdir -p "$NEW_PKG"/database/{postgres,redis,migrations}
mkdir -p "$NEW_PKG"/messaging/{kafka,nats,events}
mkdir -p "$NEW_PKG"/telemetry/{metrics,tracing,logging}
mkdir -p "$NEW_PKG"/http/{middleware,handlers,client}
mkdir -p "$NEW_PKG"/grpc/{interceptors,health,reflection}
mkdir -p "$NEW_PKG"/k8s/{client,informers,controllers}
mkdir -p "$NEW_PKG"/testing/{fixtures,mocks,integration}
mkdir -p "$NEW_PKG"/errors/{types,handlers}
mkdir -p "$NEW_PKG"/utils/{retry,circuit,pool}

# Function to migrate a package
migrate_package() {
    local src=$1
    local dst=$2
    local desc=$3
    
    if [ -d "$src" ]; then
        echo -e "${YELLOW}Migrating $desc...${NC}"
        cp -r "$src"/* "$dst/" 2>/dev/null || true
        
        # Update import paths
        find "$dst" -type f -name "*.go" -exec sed -i \
            -e "s|github.com/phoenix/phoenix-platform/pkg|github.com/phoenix-vnext/pkg|g" \
            -e "s|github.com/phoenix/|github.com/phoenix-vnext/|g" \
            {} +
        
        echo -e "${GREEN}✓ Migrated $desc${NC}"
    else
        echo -e "${RED}✗ Source not found: $src${NC}"
    fi
}

# Migrate authentication packages
echo -e "\n${BLUE}Migrating Authentication Packages${NC}"
migrate_package "$OLD_PKG/auth" "$NEW_PKG/auth" "authentication utilities"

# Migrate database packages
echo -e "\n${BLUE}Migrating Database Packages${NC}"
migrate_package "$OLD_PKG/store" "$NEW_PKG/database/postgres" "PostgreSQL store"
if [ -d "OLD_IMPLEMENTATION/phoenix-platform/migrations" ]; then
    cp -r OLD_IMPLEMENTATION/phoenix-platform/migrations/* "$NEW_PKG/database/migrations/" 2>/dev/null || true
    echo -e "${GREEN}✓ Migrated database migrations${NC}"
fi

# Migrate messaging packages
echo -e "\n${BLUE}Migrating Messaging Packages${NC}"
migrate_package "$OLD_PKG/eventbus" "$NEW_PKG/messaging/events" "event bus"

# Migrate telemetry packages
echo -e "\n${BLUE}Migrating Telemetry Packages${NC}"
migrate_package "$OLD_PKG/metrics" "$NEW_PKG/telemetry/metrics" "metrics utilities"

# Migrate HTTP/gRPC packages
echo -e "\n${BLUE}Migrating Communication Packages${NC}"
migrate_package "$OLD_PKG/clients" "$NEW_PKG/http/client" "HTTP clients"

# Copy proto files
if [ -d "OLD_IMPLEMENTATION/phoenix-platform/api/proto" ]; then
    mkdir -p "$NEW_PKG/grpc/proto"
    cp -r OLD_IMPLEMENTATION/phoenix-platform/api/proto/* "$NEW_PKG/grpc/proto/" 2>/dev/null || true
    echo -e "${GREEN}✓ Migrated protocol buffer definitions${NC}"
fi

# Migrate utility packages
echo -e "\n${BLUE}Migrating Utility Packages${NC}"
migrate_package "$OLD_PKG/utils" "$NEW_PKG/utils" "utility functions"
migrate_package "$OLD_PKG/errors" "$NEW_PKG/errors" "error handling"

# Migrate specialized packages
echo -e "\n${BLUE}Migrating Specialized Packages${NC}"
migrate_package "$OLD_PKG/analysis" "$NEW_PKG/analysis" "analysis utilities"
migrate_package "$OLD_PKG/generator" "$NEW_PKG/generator" "generator utilities"
migrate_package "$OLD_PKG/interfaces" "$NEW_PKG/interfaces" "shared interfaces"

# Create go.mod for shared packages
echo -e "\n${YELLOW}Creating go.mod for shared packages...${NC}"
cat > "$NEW_PKG/go.mod" << EOF
module github.com/phoenix-vnext/pkg

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/grpc-ecosystem/grpc-gateway/v2 v2.18.0
    github.com/prometheus/client_golang v1.17.0
    github.com/stretchr/testify v1.8.4
    go.opentelemetry.io/otel v1.19.0
    go.opentelemetry.io/otel/trace v1.19.0
    google.golang.org/grpc v1.59.0
    google.golang.org/protobuf v1.31.0
    gorm.io/driver/postgres v1.5.3
    gorm.io/gorm v1.25.5
    k8s.io/api v0.28.3
    k8s.io/apimachinery v0.28.3
    k8s.io/client-go v0.28.3
)
EOF

# Create README for pkg directory
echo -e "\n${YELLOW}Creating package documentation...${NC}"
cat > "$NEW_PKG/README.md" << EOF
# Phoenix Platform Shared Packages

This directory contains shared Go packages used across all Phoenix platform services.

## Package Structure

### Core Packages

- **auth/**: Authentication and authorization utilities
  - \`jwt/\`: JWT token handling
  - \`oauth/\`: OAuth2 integration
  - \`rbac/\`: Role-based access control

- **database/**: Database connectivity and ORM
  - \`postgres/\`: PostgreSQL utilities
  - \`redis/\`: Redis client wrappers
  - \`migrations/\`: Database migration scripts

- **messaging/**: Event-driven communication
  - \`kafka/\`: Kafka producer/consumer
  - \`nats/\`: NATS messaging
  - \`events/\`: Event definitions and handlers

### Infrastructure Packages

- **telemetry/**: Observability utilities
  - \`metrics/\`: Prometheus metrics
  - \`tracing/\`: OpenTelemetry tracing
  - \`logging/\`: Structured logging

- **http/**: HTTP utilities
  - \`middleware/\`: Common middleware
  - \`handlers/\`: Shared handlers
  - \`client/\`: HTTP client with retry

- **grpc/**: gRPC utilities
  - \`interceptors/\`: Common interceptors
  - \`health/\`: Health check implementation
  - \`reflection/\`: gRPC reflection

### Platform Packages

- **k8s/**: Kubernetes integration
  - \`client/\`: K8s client utilities
  - \`informers/\`: Resource informers
  - \`controllers/\`: Controller helpers

- **testing/**: Test utilities
  - \`fixtures/\`: Test data fixtures
  - \`mocks/\`: Mock generators
  - \`integration/\`: Integration test helpers

### Utility Packages

- **errors/**: Error handling
  - \`types/\`: Custom error types
  - \`handlers/\`: Error handlers

- **utils/**: General utilities
  - \`retry/\`: Retry mechanisms
  - \`circuit/\`: Circuit breaker
  - \`pool/\`: Resource pooling

## Usage

Import packages in your service:

\`\`\`go
import (
    "github.com/phoenix-vnext/pkg/auth/jwt"
    "github.com/phoenix-vnext/pkg/database/postgres"
    "github.com/phoenix-vnext/pkg/telemetry/metrics"
)
\`\`\`

## Development

When adding new shared functionality:

1. Determine the appropriate package
2. Add comprehensive tests
3. Document the API
4. Update this README
5. Version appropriately

## Testing

Run all package tests:

\`\`\`bash
go test ./...
\`\`\`

Run with coverage:

\`\`\`bash
go test -cover ./...
\`\`\`
EOF

# Create example implementations for key packages
echo -e "\n${YELLOW}Creating example implementations...${NC}"

# HTTP Middleware example
cat > "$NEW_PKG/http/middleware/logging.go" << 'EOF'
package middleware

import (
    "time"
    
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

// Logging middleware for HTTP requests
func Logging(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        raw := c.Request.URL.RawQuery

        // Process request
        c.Next()

        // Log request details
        latency := time.Since(start)
        clientIP := c.ClientIP()
        method := c.Request.Method
        statusCode := c.Writer.Status()

        if raw != "" {
            path = path + "?" + raw
        }

        logger.Info("HTTP Request",
            zap.String("method", method),
            zap.String("path", path),
            zap.Int("status", statusCode),
            zap.String("client_ip", clientIP),
            zap.Duration("latency", latency),
            zap.String("user_agent", c.Request.UserAgent()),
        )
    }
}
EOF

# Error types example
cat > "$NEW_PKG/errors/types/errors.go" << 'EOF'
package types

import (
    "fmt"
    "net/http"
)

// Error represents a Phoenix platform error
type Error struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Status  int    `json:"status"`
    Details any    `json:"details,omitempty"`
}

func (e *Error) Error() string {
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Common errors
var (
    ErrNotFound = &Error{
        Code:    "NOT_FOUND",
        Message: "Resource not found",
        Status:  http.StatusNotFound,
    }
    
    ErrUnauthorized = &Error{
        Code:    "UNAUTHORIZED",
        Message: "Unauthorized access",
        Status:  http.StatusUnauthorized,
    }
    
    ErrValidation = &Error{
        Code:    "VALIDATION_ERROR",
        Message: "Validation failed",
        Status:  http.StatusBadRequest,
    }
)

// New creates a new error
func New(code, message string, status int) *Error {
    return &Error{
        Code:    code,
        Message: message,
        Status:  status,
    }
}
EOF

# Count migrated files
TOTAL_FILES=$(find "$NEW_PKG" -type f -name "*.go" | wc -l)

echo -e "\n${GREEN}=== Migration Summary ===${NC}"
echo -e "Total Go files migrated: ${BLUE}$TOTAL_FILES${NC}"
echo -e "Package structure created: ${GREEN}✓${NC}"
echo -e "Import paths updated: ${GREEN}✓${NC}"
echo -e "Documentation created: ${GREEN}✓${NC}"

echo -e "\n${YELLOW}Next steps:${NC}"
echo "1. Review migrated packages in $NEW_PKG"
echo "2. Run 'cd $NEW_PKG && go mod tidy' to update dependencies"
echo "3. Run 'cd $NEW_PKG && go test ./...' to verify functionality"
echo "4. Update service imports to use new package paths"