# Phoenix Platform Makefile

.PHONY: all build test clean help

# Variables
DOCKER_REGISTRY ?= phoenix
VERSION ?= latest
GO_VERSION = 1.21
NODE_VERSION = 18
KUBEBUILDER_VERSION = 3.11.1

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOTEST = $(GOCMD) test
GOMOD = $(GOCMD) mod
GOFMT = gofmt
GOLINT = golangci-lint

# Directories
BUILD_DIR = build
DIST_DIR = dist
VENDOR_DIR = vendor

# All targets
all: build

## help: Show this help message
help:
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## docs: Build documentation
docs:
	@echo "Building documentation with MkDocs..."
	@cd .. && mkdocs build

## docs-serve: Serve documentation locally
docs-serve:
	@echo "Serving documentation at http://localhost:8000..."
	@cd .. && mkdocs serve

## docs-deploy: Deploy documentation to GitHub Pages
docs-deploy:
	@echo "Deploying documentation to GitHub Pages..."
	@cd .. && mkdocs gh-deploy

## docs-check: Run documentation quality checks
docs-check:
	@echo "Running documentation quality checks..."
	@cd .. && ./scripts/check-docs.sh

## docs-generate-api: Generate API documentation from proto files
docs-generate-api:
	@echo "Generating API documentation..."
	@cd .. && ./scripts/generate-api-docs.sh

## docs-version: Deploy a new documentation version
docs-version:
	@echo "Deploying documentation version..."
	@cd .. && ./scripts/docs-version.sh deploy $(VERSION)

## deps: Download and install dependencies
deps:
	@echo "Installing dependencies..."
	@$(GOMOD) download
	@cd dashboard && npm ci
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
	@go install sigs.k8s.io/kustomize/kustomize/v4@latest

## fmt: Format code
fmt:
	@echo "Formatting Go code..."
	@$(GOFMT) -w -s .
	@echo "Formatting frontend code..."
	@cd dashboard && npm run format

## lint: Run linters
lint:
	@echo "Linting Go code..."
	@$(GOLINT) run ./...
	@echo "Linting frontend code..."
	@cd dashboard && npm run lint

## test: Run all tests
test: test-unit test-integration

## test-unit: Run unit tests
test-unit:
	@echo "Running Go unit tests..."
	@$(GOTEST) -v -race -coverprofile=coverage.out \
		$$(go list ./... | grep -v /test/integration)
	@echo "Running Dashboard unit tests..."
	@cd dashboard && npm test

## test-integration: Run integration tests
test-integration:
	@echo "Running integration tests..."
	@cd test && ./run-integration-tests.sh

## test-dashboard: Run dashboard tests with coverage
test-dashboard:
	@echo "Running dashboard tests with coverage..."
	@cd dashboard && npm run test:coverage

## test-e2e: Run end-to-end tests
test-e2e:
	@echo "Running e2e tests..."
	@$(GOTEST) -v -tags=e2e -timeout=30m ./test/e2e/...

## build: Build all components
build: build-api build-controller build-generator build-control-service build-api-gateway build-operators build-simulator build-dashboard build-cli

## build-api: Build API service
build-api:
	@echo "Building API service..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 $(GOBUILD) -o $(BUILD_DIR)/phoenix-api ./cmd/api

## build-controller: Build experiment controller
build-controller:
	@echo "Building experiment controller..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 $(GOBUILD) -o $(BUILD_DIR)/experiment-controller ./cmd/controller

## build-generator: Build configuration generator
build-generator:
	@echo "Building configuration generator..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 $(GOBUILD) -o $(BUILD_DIR)/config-generator ./cmd/generator

## build-control-service: Build control service
build-control-service:
	@echo "Building control service..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 $(GOBUILD) -o $(BUILD_DIR)/control-service ./cmd/control-service

## build-api-gateway: Build API gateway
build-api-gateway:
	@echo "Building API gateway..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 $(GOBUILD) -o $(BUILD_DIR)/api-gateway ./cmd/api-gateway

## build-operators: Build Kubernetes operators
build-operators:
	@echo "Building pipeline operator..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 $(GOBUILD) -o $(BUILD_DIR)/pipeline-operator ./operators/pipeline/cmd
	@echo "Building loadsim operator..."
	@CGO_ENABLED=0 $(GOBUILD) -o $(BUILD_DIR)/loadsim-operator ./operators/loadsim/cmd

## build-simulator: Build process simulator
build-simulator:
	@echo "Building process simulator..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 $(GOBUILD) -o $(BUILD_DIR)/process-simulator ./cmd/simulator

## build-dashboard: Build web dashboard
build-dashboard:
	@echo "Building dashboard..."
	@cd dashboard && npm run build
	@mkdir -p $(DIST_DIR)/dashboard
	@cp -r dashboard/dist/* $(DIST_DIR)/dashboard/

## build-cli: Build Phoenix CLI
build-cli:
	@echo "Building Phoenix CLI..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 $(GOBUILD) -ldflags="-X 'github.com/phoenix-platform/cmd/phoenix-cli/cmd.Version=$(VERSION)' -X 'github.com/phoenix-platform/cmd/phoenix-cli/cmd.GitCommit=$$(git rev-parse --short HEAD)' -X 'github.com/phoenix-platform/cmd/phoenix-cli/cmd.BuildDate=$$(date -u +%Y-%m-%dT%H:%M:%SZ)'" -o $(BUILD_DIR)/phoenix ./cmd/phoenix-cli

## docker: Build all Docker images
docker: docker-api docker-controller docker-generator docker-control-service docker-api-gateway docker-operators docker-simulator docker-dashboard

## docker-api: Build API Docker image
docker-api:
	@echo "Building API Docker image..."
	@docker build -f docker/api/Dockerfile -t $(DOCKER_REGISTRY)/phoenix-api:$(VERSION) .

## docker-controller: Build controller Docker image
docker-controller:
	@echo "Building controller Docker image..."
	@docker build -f docker/controller/Dockerfile -t $(DOCKER_REGISTRY)/experiment-controller:$(VERSION) .

## docker-generator: Build generator Docker image
docker-generator:
	@echo "Building generator Docker image..."
	@docker build -f docker/generator/Dockerfile -t $(DOCKER_REGISTRY)/config-generator:$(VERSION) .

## docker-control-service: Build control service Docker image
docker-control-service:
	@echo "Building control service Docker image..."
	@docker build -f cmd/control-service/Dockerfile -t $(DOCKER_REGISTRY)/control-service:$(VERSION) .

## docker-api-gateway: Build API gateway Docker image
docker-api-gateway:
	@echo "Building API gateway Docker image..."
	@docker build -f cmd/api-gateway/Dockerfile -t $(DOCKER_REGISTRY)/api-gateway:$(VERSION) .

## docker-operators: Build operator Docker images
docker-operators:
	@echo "Building operator Docker images..."
	@docker build -f docker/pipeline-operator/Dockerfile -t $(DOCKER_REGISTRY)/pipeline-operator:$(VERSION) .
	@docker build -f docker/loadsim-operator/Dockerfile -t $(DOCKER_REGISTRY)/loadsim-operator:$(VERSION) .

## docker-simulator: Build simulator Docker image
docker-simulator:
	@echo "Building simulator Docker image..."
	@docker build -f docker/simulator/Dockerfile -t $(DOCKER_REGISTRY)/process-simulator:$(VERSION) .

## docker-dashboard: Build dashboard Docker image
docker-dashboard: build-dashboard
	@echo "Building dashboard Docker image..."
	@docker build -f docker/dashboard/Dockerfile -t $(DOCKER_REGISTRY)/phoenix-dashboard:$(VERSION) .

## push: Push all Docker images
push:
	@echo "Pushing Docker images..."
	@docker push $(DOCKER_REGISTRY)/phoenix-api:$(VERSION)
	@docker push $(DOCKER_REGISTRY)/experiment-controller:$(VERSION)
	@docker push $(DOCKER_REGISTRY)/config-generator:$(VERSION)
	@docker push $(DOCKER_REGISTRY)/pipeline-operator:$(VERSION)
	@docker push $(DOCKER_REGISTRY)/loadsim-operator:$(VERSION)
	@docker push $(DOCKER_REGISTRY)/process-simulator:$(VERSION)
	@docker push $(DOCKER_REGISTRY)/phoenix-dashboard:$(VERSION)

## generate: Generate code and manifests
generate: generate-proto generate-crds
	@echo "Code generation complete"

## generate-proto: Generate protobuf code
generate-proto:
	@echo "Generating protobuf code..."
	@if ! command -v protoc >/dev/null 2>&1; then \
		echo "Error: protoc not installed. Install with: brew install protobuf"; \
		exit 1; \
	fi
	@if ! command -v protoc-gen-go >/dev/null 2>&1; then \
		echo "Installing protoc-gen-go..."; \
		go install google.golang.org/protobuf/cmd/protoc-gen-go@latest; \
	fi
	@if ! command -v protoc-gen-go-grpc >/dev/null 2>&1; then \
		echo "Installing protoc-gen-go-grpc..."; \
		go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest; \
	fi
	@./scripts/generate-proto.sh

## generate-crds: Generate Kubernetes CRDs
generate-crds:
	@echo "Generating CRDs..."
	@controller-gen crd paths="./operators/..." output:crd:artifacts:config=k8s/crds
	@echo "Generating Go code..."
	@controller-gen object paths="./operators/..."

## manifests: Generate Kubernetes manifests
manifests: generate
	@echo "Building Kubernetes manifests..."
	@kustomize build k8s/base > k8s/phoenix-all.yaml

## install-crds: Install CRDs in cluster
install-crds:
	@echo "Installing CRDs..."
	@kubectl apply -f k8s/crds/

## deploy: Deploy Phoenix to Kubernetes
deploy: install-crds
	@echo "Deploying Phoenix..."
	@kubectl apply -k k8s/overlays/development

## undeploy: Remove Phoenix from Kubernetes
undeploy:
	@echo "Removing Phoenix..."
	@kubectl delete -k k8s/overlays/development
	@kubectl delete -f k8s/crds/

## cluster-up: Start local Kubernetes cluster
cluster-up:
	@echo "Starting kind cluster..."
	@kind create cluster --name phoenix-dev --config test/kind-config.yaml
	@kubectl cluster-info --context kind-phoenix-dev

## cluster-down: Stop local Kubernetes cluster
cluster-down:
	@echo "Deleting kind cluster..."
	@kind delete cluster --name phoenix-dev

## port-forward: Forward ports for local development
port-forward:
	@echo "Port forwarding services..."
	@kubectl port-forward -n phoenix-system svc/phoenix-api-gateway 8443:443 &
	@kubectl port-forward -n phoenix-system svc/prometheus-server 9090:80 &
	@kubectl port-forward -n phoenix-system svc/grafana 3000:80 &

## clean: Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR) $(VENDOR_DIR)
	@rm -f coverage.out coverage.html
	@cd dashboard && rm -rf dist node_modules

## setup-hooks: Setup git hooks
setup-hooks:
	@echo "Setting up git hooks..."
	@pre-commit install
	@pre-commit install --hook-type commit-msg
	@echo "Git hooks installed successfully"

## dev: Start local development environment
dev:
	@echo "Starting local development environment..."
	@./scripts/dev-environment.sh up

## dev-down: Stop local development environment
dev-down:
	@echo "Stopping local development environment..."
	@./scripts/dev-environment.sh down

## dev-logs: Show logs for development environment
dev-logs:
	@./scripts/dev-environment.sh logs-f

## dev-status: Show status of development environment
dev-status:
	@./scripts/dev-environment.sh status

## validate: Run all validation checks
validate: validate-structure validate-imports lint test

## validate-structure: Validate mono-repo structure
validate-structure:
	@echo "Validating mono-repo structure..."
	@bash scripts/validate-structure.sh

## validate-imports: Validate Go import rules
validate-imports:
	@echo "Validating Go imports..."
	@go run scripts/validate-imports.go

## verify: Run all pre-commit checks
verify: fmt lint validate test
	@echo "All checks passed ✅"

## coverage: Generate test coverage report
coverage:
	@echo "Generating test coverage report..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.DEFAULT_GOAL := help