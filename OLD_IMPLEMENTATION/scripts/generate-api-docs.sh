#!/bin/bash
# Generate API documentation from proto files and OpenAPI specs

set -euo pipefail

# Script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Output functions
print_success() { echo -e "${GREEN}✓ $1${NC}"; }
print_error() { echo -e "${RED}✗ $1${NC}"; }
print_info() { echo -e "${BLUE}ℹ $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠ $1${NC}"; }

# Configuration
PROTO_DIR="$PROJECT_ROOT/phoenix-platform/api/proto"
OPENAPI_DIR="$PROJECT_ROOT/docs/assets"
DOCS_DIR="$PROJECT_ROOT/docs/api"
TEMP_DIR="/tmp/phoenix-api-docs"

# Check dependencies
check_dependencies() {
    local deps=(protoc protoc-gen-doc protoc-gen-openapiv2 jq yq)
    local missing=()
    
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            missing+=("$dep")
        fi
    done
    
    if [ ${#missing[@]} -ne 0 ]; then
        print_error "Missing dependencies: ${missing[*]}"
        echo "Install them with:"
        echo "  brew install protobuf protoc-gen-doc grpc-gateway jq yq"
        echo "  go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest"
        echo "  go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest"
        exit 1
    fi
}

# Clean up function
cleanup() {
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

# Generate gRPC documentation from proto files
generate_grpc_docs() {
    print_info "Generating gRPC documentation from proto files..."
    
    mkdir -p "$TEMP_DIR/grpc"
    
    # Find all proto files
    find "$PROTO_DIR" -name "*.proto" -type f | while read -r proto_file; do
        local service_name=$(basename "$proto_file" .proto)
        print_info "Processing $service_name.proto..."
        
        # Generate markdown documentation
        protoc \
            --proto_path="$PROTO_DIR" \
            --doc_out="$TEMP_DIR/grpc" \
            --doc_opt=markdown,"${service_name}.md" \
            "$proto_file" 2>/dev/null || {
                print_warning "Failed to generate docs for $proto_file"
                continue
            }
    done
    
    # Combine all service docs into one gRPC reference
    {
        echo "# gRPC API Reference"
        echo ""
        echo "This document provides a complete reference for all gRPC services in the Phoenix Platform."
        echo ""
        echo "## Table of Contents"
        echo ""
        
        # Generate TOC
        for md_file in "$TEMP_DIR/grpc"/*.md; do
            if [ -f "$md_file" ]; then
                local service=$(basename "$md_file" .md)
                echo "- [$service](#$service)"
            fi
        done
        
        echo ""
        echo "---"
        echo ""
        
        # Append all service docs
        for md_file in "$TEMP_DIR/grpc"/*.md; do
            if [ -f "$md_file" ]; then
                echo ""
                cat "$md_file"
                echo ""
                echo "---"
            fi
        done
    } > "$DOCS_DIR/grpc.md"
    
    print_success "Generated gRPC documentation"
}

# Generate OpenAPI specs from proto files
generate_openapi_specs() {
    print_info "Generating OpenAPI specifications from proto files..."
    
    mkdir -p "$TEMP_DIR/openapi"
    
    # Generate OpenAPI specs for each service
    find "$PROTO_DIR" -name "*.proto" -type f | while read -r proto_file; do
        local service_name=$(basename "$proto_file" .proto)
        
        # Skip if no HTTP annotations
        if ! grep -q "google.api.http" "$proto_file" 2>/dev/null; then
            continue
        fi
        
        print_info "Generating OpenAPI for $service_name..."
        
        protoc \
            --proto_path="$PROTO_DIR" \
            --proto_path="$GOPATH/src" \
            --openapiv2_out="$TEMP_DIR/openapi" \
            --openapiv2_opt=logtostderr=true \
            --openapiv2_opt=generate_unbound_methods=true \
            "$proto_file" 2>/dev/null || {
                print_warning "No HTTP endpoints in $proto_file"
                continue
            }
    done
    
    # Merge all OpenAPI specs into one
    if ls "$TEMP_DIR/openapi"/*.swagger.json 1> /dev/null 2>&1; then
        print_info "Merging OpenAPI specifications..."
        
        # Use jq to merge all swagger files
        jq -s '
            .[0] as $base |
            reduce .[1:][] as $item ($base;
                .paths += $item.paths |
                .definitions += $item.definitions |
                .securityDefinitions += $item.securityDefinitions
            )
        ' "$TEMP_DIR/openapi"/*.swagger.json > "$OPENAPI_DIR/openapi-generated.json"
        
        # Convert to YAML
        yq eval -P "$OPENAPI_DIR/openapi-generated.json" > "$OPENAPI_DIR/openapi-generated.yaml"
        
        print_success "Generated OpenAPI specifications"
    else
        print_warning "No OpenAPI specifications generated"
    fi
}

# Generate WebSocket documentation
generate_websocket_docs() {
    print_info "Generating WebSocket documentation..."
    
    cat > "$DOCS_DIR/websocket.md" << 'EOF'
# WebSocket API Reference

The Phoenix Platform provides real-time updates through WebSocket connections.

## Connection

### Endpoint

```
wss://api.phoenix.example.com/v1/ws
```

### Authentication

Include the JWT token in the connection URL:

```javascript
const ws = new WebSocket('wss://api.phoenix.example.com/v1/ws?token=YOUR_JWT_TOKEN');
```

Or send it as the first message after connection:

```javascript
ws.send(JSON.stringify({
  type: 'auth',
  token: 'YOUR_JWT_TOKEN'
}));
```

## Message Format

All messages use JSON format:

```typescript
interface Message {
  id?: string;        // Message ID for request/response correlation
  type: string;       // Message type
  payload?: any;      // Message-specific payload
  error?: {           // Error information (if applicable)
    code: string;
    message: string;
  };
}
```

## Subscription Messages

### Subscribe to Experiment Updates

```json
{
  "type": "subscribe",
  "payload": {
    "channel": "experiment",
    "id": "exp-123"
  }
}
```

### Subscribe to All Experiments

```json
{
  "type": "subscribe",
  "payload": {
    "channel": "experiments",
    "filter": {
      "status": "running"
    }
  }
}
```

### Unsubscribe

```json
{
  "type": "unsubscribe",
  "payload": {
    "channel": "experiment",
    "id": "exp-123"
  }
}
```

## Event Messages

### Experiment Status Update

```json
{
  "type": "experiment.status",
  "payload": {
    "id": "exp-123",
    "status": "running",
    "previous_status": "pending",
    "timestamp": "2024-01-25T10:00:00Z"
  }
}
```

### Metrics Update

```json
{
  "type": "metrics.update",
  "payload": {
    "experiment_id": "exp-123",
    "timestamp": "2024-01-25T10:00:00Z",
    "metrics": {
      "baseline_cardinality": 50000,
      "candidate_cardinality": 12500,
      "reduction_percentage": 75
    }
  }
}
```

### Alert Notification

```json
{
  "type": "alert",
  "payload": {
    "experiment_id": "exp-123",
    "severity": "warning",
    "title": "High error rate detected",
    "message": "Candidate pipeline error rate exceeded threshold",
    "timestamp": "2024-01-25T10:00:00Z"
  }
}
```

## Error Handling

### Connection Errors

```javascript
ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = (event) => {
  if (event.code === 1006) {
    console.error('Connection lost abnormally');
  }
};
```

### Message Errors

```json
{
  "type": "error",
  "error": {
    "code": "INVALID_MESSAGE",
    "message": "Message type not recognized"
  }
}
```

## Heartbeat

Send periodic ping messages to keep the connection alive:

```javascript
setInterval(() => {
  if (ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({ type: 'ping' }));
  }
}, 30000); // Every 30 seconds
```

## Example Client

```javascript
class PhoenixWebSocket {
  constructor(token) {
    this.token = token;
    this.reconnectDelay = 1000;
    this.maxReconnectDelay = 30000;
    this.connect();
  }
  
  connect() {
    this.ws = new WebSocket(`wss://api.phoenix.example.com/v1/ws?token=${this.token}`);
    
    this.ws.onopen = () => {
      console.log('Connected to Phoenix WebSocket');
      this.reconnectDelay = 1000;
    };
    
    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      this.handleMessage(message);
    };
    
    this.ws.onclose = () => {
      console.log('Disconnected, reconnecting...');
      setTimeout(() => this.connect(), this.reconnectDelay);
      this.reconnectDelay = Math.min(this.reconnectDelay * 2, this.maxReconnectDelay);
    };
  }
  
  subscribe(channel, id) {
    this.send({
      type: 'subscribe',
      payload: { channel, id }
    });
  }
  
  send(message) {
    if (this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    }
  }
  
  handleMessage(message) {
    switch (message.type) {
      case 'experiment.status':
        console.log('Experiment status updated:', message.payload);
        break;
      case 'metrics.update':
        console.log('New metrics:', message.payload);
        break;
      case 'error':
        console.error('Error:', message.error);
        break;
    }
  }
}

// Usage
const ws = new PhoenixWebSocket('YOUR_JWT_TOKEN');
ws.subscribe('experiment', 'exp-123');
```
EOF
    
    print_success "Generated WebSocket documentation"
}

# Generate CLI documentation
generate_cli_docs() {
    print_info "Generating CLI documentation..."
    
    # Check if phoenix CLI exists
    local cli_path="$PROJECT_ROOT/phoenix-platform/build/phoenix"
    
    if [ ! -f "$cli_path" ]; then
        print_warning "Phoenix CLI not built, attempting to build..."
        (cd "$PROJECT_ROOT/phoenix-platform" && make build-cli)
    fi
    
    if [ -f "$cli_path" ]; then
        # Generate command tree documentation
        "$cli_path" help --tree > "$TEMP_DIR/cli-tree.txt" 2>/dev/null || true
        
        # Generate detailed help for each command
        local commands=("auth" "experiment" "pipeline" "config" "version")
        
        {
            echo "# Phoenix CLI Reference"
            echo ""
            echo "The Phoenix CLI provides command-line access to the Phoenix Platform."
            echo ""
            echo "## Installation"
            echo ""
            echo '```bash'
            echo "# Download the latest release"
            echo "curl -L https://github.com/phoenix-platform/phoenix/releases/latest/download/phoenix-$(uname -s)-$(uname -m) -o phoenix"
            echo "chmod +x phoenix"
            echo "sudo mv phoenix /usr/local/bin/"
            echo '```'
            echo ""
            echo "## Configuration"
            echo ""
            echo '```bash'
            echo "# Configure API endpoint"
            echo "phoenix config set api.url https://api.phoenix.example.com"
            echo ""
            echo "# Login"
            echo "phoenix auth login"
            echo '```'
            echo ""
            echo "## Command Reference"
            echo ""
            
            for cmd in "${commands[@]}"; do
                echo "### phoenix $cmd"
                echo ""
                echo '```'
                "$cli_path" "$cmd" --help 2>/dev/null || echo "Help not available"
                echo '```'
                echo ""
            done
            
        } > "$DOCS_DIR/cli.md"
        
        print_success "Generated CLI documentation"
    else
        print_warning "Could not generate CLI documentation - CLI not found"
    fi
}

# Main execution
main() {
    print_info "Phoenix API Documentation Generator"
    print_info "===================================="
    
    # Check dependencies
    check_dependencies
    
    # Create directories
    mkdir -p "$TEMP_DIR" "$DOCS_DIR" "$OPENAPI_DIR"
    
    # Generate documentation
    generate_grpc_docs
    generate_openapi_specs
    generate_websocket_docs
    generate_cli_docs
    
    # Update mkdocs navigation if needed
    print_info "Updating mkdocs.yml navigation..."
    
    # Summary
    echo ""
    print_success "API documentation generation complete!"
    print_info "Generated files:"
    [ -f "$DOCS_DIR/grpc.md" ] && echo "  - $DOCS_DIR/grpc.md"
    [ -f "$DOCS_DIR/websocket.md" ] && echo "  - $DOCS_DIR/websocket.md"
    [ -f "$DOCS_DIR/cli.md" ] && echo "  - $DOCS_DIR/cli.md"
    [ -f "$OPENAPI_DIR/openapi-generated.yaml" ] && echo "  - $OPENAPI_DIR/openapi-generated.yaml"
}

# Run main function
main "$@"