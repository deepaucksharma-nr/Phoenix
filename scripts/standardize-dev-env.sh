#!/bin/bash
# standardize-dev-env.sh - Standardize development environment for Phoenix Platform
# Created by Abhinav as part of developer environment standardization task

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo -e "${BLUE}=== Phoenix Platform Developer Environment Standardization ===${NC}"
echo ""

# Required versions
GO_VERSION_MIN="1.21"
NODE_VERSION_MIN="18.0.0"
DOCKER_VERSION_MIN="20.10.0"
DOCKER_COMPOSE_VERSION_MIN="2.0.0"

# Helper functions
function version_gt() { 
    test "$(printf '%s\n' "$@" | sort -V | head -n 1)" != "$1"
}

function check_command() {
    if command -v "$1" > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

function check_version() {
    local command="$1"
    local current="$2"
    local required="$3"
    local name="$4"

    if version_gt "$current" "$required" || [ "$current" = "$required" ]; then
        echo -e "${GREEN}✓${NC} $name $current installed (minimum: $required)"
        return 0
    else
        echo -e "${RED}✗${NC} $name $current is outdated (minimum: $required)"
        return 1
    fi
}

# Check prerequisites
function check_prerequisites() {
    echo -e "${YELLOW}Checking prerequisites...${NC}"
    local all_passed=true

    # Check Go
    if check_command "go"; then
        GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        check_version "go" "$GO_VERSION" "$GO_VERSION_MIN" "Go" || all_passed=false
    else
        echo -e "${RED}✗${NC} Go is not installed (minimum: $GO_VERSION_MIN)"
        echo "   Install from: https://golang.org/doc/install"
        all_passed=false
    fi

    # Check Node.js
    if check_command "node"; then
        NODE_VERSION=$(node -v | sed 's/v//')
        check_version "node" "$NODE_VERSION" "$NODE_VERSION_MIN" "Node.js" || all_passed=false
    else
        echo -e "${RED}✗${NC} Node.js is not installed (minimum: $NODE_VERSION_MIN)"
        echo "   Install from: https://nodejs.org/"
        all_passed=false
    fi

    # Check npm
    if ! check_command "npm"; then
        echo -e "${RED}✗${NC} npm is not installed"
        echo "   It should be installed with Node.js"
        all_passed=false
    fi

    # Check pnpm
    if ! check_command "pnpm"; then
        echo -e "${YELLOW}!${NC} pnpm is not installed, will install later"
    fi

    # Check Docker
    if check_command "docker"; then
        DOCKER_VERSION=$(docker --version | sed -E 's/.*version ([0-9]+\.[0-9]+\.[0-9]+).*/\1/')
        check_version "docker" "$DOCKER_VERSION" "$DOCKER_VERSION_MIN" "Docker" || all_passed=false
    else
        echo -e "${RED}✗${NC} Docker is not installed (minimum: $DOCKER_VERSION_MIN)"
        echo "   Install from: https://docs.docker.com/get-docker/"
        all_passed=false
    fi

    # Check Docker Compose
    if docker compose version > /dev/null 2>&1; then
        DOCKER_COMPOSE_VERSION=$(docker compose version | grep -o '[0-9]\+\.[0-9]\+\.[0-9]\+' | head -1)
        check_version "docker compose" "$DOCKER_COMPOSE_VERSION" "$DOCKER_COMPOSE_VERSION_MIN" "Docker Compose" || all_passed=false
    elif check_command "docker-compose"; then
        DOCKER_COMPOSE_VERSION=$(docker-compose --version | grep -o '[0-9]\+\.[0-9]\+\.[0-9]\+' | head -1)
        check_version "docker-compose" "$DOCKER_COMPOSE_VERSION" "$DOCKER_COMPOSE_VERSION_MIN" "Docker Compose" || all_passed=false
    else
        echo -e "${RED}✗${NC} Docker Compose is not installed (minimum: $DOCKER_COMPOSE_VERSION_MIN)"
        echo "   Install from: https://docs.docker.com/compose/install/"
        all_passed=false
    fi

    # Check Make
    if ! check_command "make"; then
        echo -e "${RED}✗${NC} Make is not installed"
        echo "   Install via your OS package manager"
        all_passed=false
    fi

    # Check Git
    if ! check_command "git"; then
        echo -e "${RED}✗${NC} Git is not installed"
        echo "   Install from: https://git-scm.com/downloads"
        all_passed=false
    fi
    
    return $([ "$all_passed" = true ] && echo 0 || echo 1)
}

function install_required_tools() {
    echo -e "\n${YELLOW}Installing required development tools...${NC}"

    # Install pnpm if needed
    if ! check_command "pnpm"; then
        echo "Installing pnpm package manager..."
        npm install -g pnpm
        echo -e "${GREEN}✓${NC} pnpm installed"
    fi

    # Install Go development tools
    echo "Installing Go development tools..."
    
    # Create tools directory if it doesn't exist
    if [ ! -d "$REPO_ROOT/tools/bin" ]; then
        mkdir -p "$REPO_ROOT/tools/bin"
    fi
    
    # Check if golangci-lint is installed
    if ! check_command "golangci-lint"; then
        echo "Installing golangci-lint..."
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$REPO_ROOT/tools/bin" v1.55.2
        echo -e "${GREEN}✓${NC} golangci-lint installed"
    fi
    
    # Install Go tools
    GOTOOLS=(
        "github.com/golang/mock/mockgen@v1.6.0"
        "github.com/google/wire/cmd/wire@latest"
        "github.com/swaggo/swag/cmd/swag@v1.16.2"
        "mvdan.cc/gofumpt@latest"
        "github.com/bufbuild/buf/cmd/buf@v1.28.1"
        "github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    )
    
    for tool in "${GOTOOLS[@]}"; do
        name=$(echo "$tool" | cut -d'/' -f3 | cut -d'@' -f1)
        echo "Installing $name..."
        go install "$tool"
    done
    
    echo -e "${GREEN}✓${NC} Go tools installed"

    # Install Node.js tools
    echo "Installing Node.js tools..."
    npm_tools=(
        "eslint"
        "prettier"
        "@commitlint/cli"
        "@commitlint/config-conventional"
    )
    
    npm install -g "${npm_tools[@]}"
    echo -e "${GREEN}✓${NC} Node.js tools installed"

    # Install pre-commit hooks
    echo "Setting up pre-commit hooks..."
    if ! check_command "pre-commit"; then
        if check_command "pip3"; then
            pip3 install pre-commit
        elif check_command "pip"; then
            pip install pre-commit
        else
            echo -e "${YELLOW}!${NC} pip not found, skipping pre-commit installation"
            echo "   Install Python and pip then run: pip install pre-commit"
        fi
    fi
    
    # Initialize pre-commit if .pre-commit-config.yaml exists
    if [ -f "$REPO_ROOT/.pre-commit-config.yaml" ]; then
        cd "$REPO_ROOT" && pre-commit install --install-hooks
        pre-commit install --hook-type commit-msg
        echo -e "${GREEN}✓${NC} pre-commit hooks installed"
    else
        echo -e "${YELLOW}!${NC} .pre-commit-config.yaml not found, skipping hook installation"
    fi
}

function configure_environments() {
    echo -e "\n${YELLOW}Configuring development environments...${NC}"

    # Create .env file if it doesn't exist
    if [ ! -f "$REPO_ROOT/.env" ]; then
        echo "Creating .env file from template..."
        if [ -f "$REPO_ROOT/.env.template" ]; then
            cp "$REPO_ROOT/.env.template" "$REPO_ROOT/.env"
            echo -e "${GREEN}✓${NC} .env file created"
        else
            echo -e "${YELLOW}!${NC} .env.template not found, skipping .env creation"
        fi
    else
        echo -e "${GREEN}✓${NC} .env file already exists"
    fi

    # Setup Visual Studio Code settings
    if [ ! -d "$REPO_ROOT/.vscode" ]; then
        echo "Setting up VS Code configuration..."
        mkdir -p "$REPO_ROOT/.vscode"
        
        # settings.json
        cat > "$REPO_ROOT/.vscode/settings.json" << 'EOF'
{
    "go.useLanguageServer": true,
    "go.lintTool": "golangci-lint",
    "go.lintOnSave": "workspace",
    "go.formatTool": "goimports",
    "go.formatFlags": ["-local", "github.com/phoenix/platform"],
    "go.testTimeout": "10s",
    "go.coverOnSave": true,
    "editor.formatOnSave": true,
    "[go]": {
        "editor.codeActionsOnSave": {
            "source.organizeImports": "explicit"
        },
        "editor.defaultFormatter": "golang.go"
    },
    "[javascript]": {
        "editor.defaultFormatter": "esbenp.prettier-vscode",
        "editor.formatOnSave": true
    },
    "[typescript]": {
        "editor.defaultFormatter": "esbenp.prettier-vscode",
        "editor.formatOnSave": true
    },
    "[typescriptreact]": {
        "editor.defaultFormatter": "esbenp.prettier-vscode",
        "editor.formatOnSave": true
    },
    "[json]": {
        "editor.defaultFormatter": "esbenp.prettier-vscode",
        "editor.formatOnSave": true
    },
    "[yaml]": {
        "editor.defaultFormatter": "esbenp.prettier-vscode",
        "editor.formatOnSave": true
    },
    "files.insertFinalNewline": true,
    "files.trimTrailingWhitespace": true,
    "editor.tabSize": 2,
    "editor.detectIndentation": false,
    "files.exclude": {
        "**/.git": true,
        "**/.DS_Store": true,
        "**/node_modules": true,
        "data/": true,
        "tmp/": true
    },
    "search.exclude": {
        "**/node_modules": true,
        "data/": true,
        "tmp/": true
    },
    "extensions.ignoreRecommendations": false
}
EOF
        
        # launch.json
        cat > "$REPO_ROOT/.vscode/launch.json" << 'EOF'
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch API",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/projects/platform-api/cmd/api",
            "env": {
                "LOG_LEVEL": "debug"
            },
            "args": []
        },
        {
            "name": "Launch Controller",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/projects/controller/cmd",
            "env": {
                "LOG_LEVEL": "debug"
            },
            "args": []
        },
        {
            "name": "Launch Dashboard (Node)",
            "type": "node",
            "request": "launch",
            "cwd": "${workspaceFolder}/projects/dashboard",
            "runtimeExecutable": "npm",
            "runtimeArgs": ["run", "dev"],
            "env": {
                "LOG_LEVEL": "debug"
            }
        }
    ]
}
EOF

        # extensions.json
        cat > "$REPO_ROOT/.vscode/extensions.json" << 'EOF'
{
    "recommendations": [
        "golang.go",
        "davidanson.vscode-markdownlint",
        "esbenp.prettier-vscode",
        "dbaeumer.vscode-eslint",
        "ms-azuretools.vscode-docker",
        "redhat.vscode-yaml",
        "timonwong.shellcheck",
        "foxundermoon.shell-format",
        "hashicorp.terraform",
        "zxh404.vscode-proto3",
        "streetsidesoftware.code-spell-checker",
        "yoavbls.pretty-ts-errors",
        "ms-kubernetes-tools.vscode-kubernetes-tools"
    ]
}
EOF
        echo -e "${GREEN}✓${NC} VS Code configuration created"
    else
        echo -e "${GREEN}✓${NC} VS Code configuration already exists"
    fi

    # Setup Git hooks
    if [ -d "$REPO_ROOT/.git" ] && [ ! -f "$REPO_ROOT/.git/hooks/pre-commit" ]; then
        echo "Setting up Git hooks..."
        
        # pre-commit hook
        cat > "$REPO_ROOT/.git/hooks/pre-commit" << 'EOF'
#!/bin/bash
echo "Running pre-commit checks..."

# Check if Go files are formatted correctly
unformatted_go_files=$(gofmt -l .)
if [[ -n "$unformatted_go_files" ]]; then
    echo "❌ The following Go files need formatting:"
    echo "$unformatted_go_files"
    echo "Run: gofmt -w ."
    exit 1
fi

# Run golangci-lint if available
if command -v golangci-lint > /dev/null 2>&1; then
    echo "Running Go linters..."
    if ! golangci-lint run --fast; then
        echo "❌ Go linting failed!"
        exit 1
    fi
fi

# Run JavaScript/TypeScript linters if files were changed
js_ts_files=$(git diff --cached --name-only | grep -E '\.(js|jsx|ts|tsx)$')
if [[ -n "$js_ts_files" ]] && [[ -f package.json ]]; then
    echo "Running JavaScript/TypeScript linters..."
    if ! npm run lint > /dev/null 2>&1; then
        echo "❌ JavaScript/TypeScript linting failed!"
        echo "Run: npm run lint"
        exit 1
    fi
fi

echo "✅ Pre-commit checks passed"
exit 0
EOF
        chmod +x "$REPO_ROOT/.git/hooks/pre-commit"
        
        # commit-msg hook for conventional commits
        cat > "$REPO_ROOT/.git/hooks/commit-msg" << 'EOF'
#!/bin/bash
# Enforce conventional commit messages

commit_msg_file=$1
commit_msg=$(cat "$commit_msg_file")
commit_pattern='^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-z0-9-]+\))?: .{1,100}'

if ! [[ "$commit_msg" =~ $commit_pattern ]]; then
    echo "❌ Invalid commit message format."
    echo "Please follow the conventional commit format:"
    echo "  type(scope): subject"
    echo ""
    echo "Examples:"
    echo "  feat(api): add new user endpoint"
    echo "  fix(dashboard): resolve layout issue"
    echo "  docs: update README installation steps"
    exit 1
fi

exit 0
EOF
        chmod +x "$REPO_ROOT/.git/hooks/commit-msg"
        echo -e "${GREEN}✓${NC} Git hooks configured"
    elif [ -f "$REPO_ROOT/.git/hooks/pre-commit" ]; then
        echo -e "${GREEN}✓${NC} Git hooks already configured"
    fi

    # Create or update Makefiles for standardization
    if [ ! -f "$REPO_ROOT/build/makefiles/common.mk" ]; then
        echo "Creating standard Makefiles..."
        mkdir -p "$REPO_ROOT/build/makefiles"
        
        # common.mk
        cat > "$REPO_ROOT/build/makefiles/common.mk" << 'EOF'
# Phoenix Platform - Common Makefile Configuration

# Default shell
SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

# Colors for output
GREEN := \033[0;32m
RED := \033[0;31m
YELLOW := \033[0;33m
CYAN := \033[0;36m
NC := \033[0m

# Common variables
ROOT_DIR := $(shell git rev-parse --show-toplevel 2>/dev/null || pwd)
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION := $(shell cat $(ROOT_DIR)/VERSION 2>/dev/null || echo "0.1.0")
BUILD_DIR := $(ROOT_DIR)/build
TOOLS_DIR := $(ROOT_DIR)/tools
BIN_DIR := $(ROOT_DIR)/bin
PROJECTS_DIR := $(ROOT_DIR)/projects

# Helper functions
define log_info
	@echo -e "$(CYAN)[INFO]$(NC) $(1)"
endef

define log_success
	@echo -e "$(GREEN)[SUCCESS]$(NC) $(1)"
endef

define log_error
	@echo -e "$(RED)[ERROR]$(NC) $(1)"
endef

define log_warn
	@echo -e "$(YELLOW)[WARNING]$(NC) $(1)"
endef

# Common targets
.PHONY: help
help:
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
EOF

        # go.mk
        cat > "$REPO_ROOT/build/makefiles/go.mk" << 'EOF'
# Phoenix Platform - Go Project Makefile

include $(ROOT_DIR)/build/makefiles/common.mk

# Go variables
GO_FILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*" -not -path "*/mock/*" -not -path "*/generated/*")
GO_DIRS := $(shell go list -f '{{.Dir}}' ./... 2>/dev/null || echo ".")
GO_PKGS := $(shell go list ./... 2>/dev/null || echo ".")
GO_MAIN_DIRS := $(shell find . -type f -name 'main.go' -not -path "./vendor/*" | xargs -I{} dirname {})

# Build flags
LD_FLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(GIT_COMMIT) -X main.buildDate=$(BUILD_DATE)"

# Go targets
.PHONY: build

build: ## Build Go binary
	$(call log_info,"Building $(SERVICE_NAME)...")
	@for dir in $(GO_MAIN_DIRS); do \
		out_file=$$(basename $$dir); \
		echo "Building $$dir -> $(BIN_DIR)/$$out_file"; \
		go build $(LD_FLAGS) -o $(BIN_DIR)/$$out_file $$dir; \
	done
	$(call log_success,"Build complete!")

.PHONY: test
test: ## Run Go tests
	$(call log_info,"Running tests...")
	@go test -v ./...
	$(call log_success,"Tests complete!")

.PHONY: test-coverage
test-coverage: ## Run Go tests with coverage
	$(call log_info,"Running tests with coverage...")
	@mkdir -p $(ROOT_DIR)/tmp/coverage
	@go test -v -coverprofile=$(ROOT_DIR)/tmp/coverage/$(SERVICE_NAME).out ./...
	@go tool cover -html=$(ROOT_DIR)/tmp/coverage/$(SERVICE_NAME).out -o $(ROOT_DIR)/tmp/coverage/$(SERVICE_NAME).html
	$(call log_success,"Coverage report generated at tmp/coverage/$(SERVICE_NAME).html")

.PHONY: lint
lint: ## Run Go linters
	$(call log_info,"Running linters...")
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout 3m ./...; \
	else \
		go vet ./...; \
	fi
	$(call log_success,"Linting complete!")

.PHONY: fmt
fmt: ## Format Go code
	$(call log_info,"Formatting code...")
	@gofmt -w $(GO_FILES)
	$(call log_success,"Formatting complete!")

.PHONY: clean
clean: ## Clean build artifacts
	$(call log_info,"Cleaning build artifacts...")
	@rm -rf $(BIN_DIR)/$(SERVICE_NAME) $(ROOT_DIR)/tmp/coverage/$(SERVICE_NAME).*
	$(call log_success,"Clean complete!")

.PHONY: vendor
vendor: ## Update Go dependencies
	$(call log_info,"Updating dependencies...")
	@go mod tidy
	@go mod vendor
	$(call log_success,"Dependencies updated!")
EOF

        # node.mk
        cat > "$REPO_ROOT/build/makefiles/node.mk" << 'EOF'
# Phoenix Platform - Node.js Project Makefile

include $(ROOT_DIR)/build/makefiles/common.mk

# Node variables
PACKAGE_MANAGER := $(shell [ -f "pnpm-lock.yaml" ] && echo "pnpm" || ([ -f "yarn.lock" ] && echo "yarn" || echo "npm"))

# Define commands based on package manager
ifeq ($(PACKAGE_MANAGER),pnpm)
  INSTALL_CMD := pnpm install
  BUILD_CMD := pnpm build
  TEST_CMD := pnpm test
  LINT_CMD := pnpm lint
  DEV_CMD := pnpm dev
  CLEAN_CMD := pnpm clean
else ifeq ($(PACKAGE_MANAGER),yarn)
  INSTALL_CMD := yarn install --frozen-lockfile
  BUILD_CMD := yarn build
  TEST_CMD := yarn test
  LINT_CMD := yarn lint
  DEV_CMD := yarn dev
  CLEAN_CMD := yarn clean
else
  INSTALL_CMD := npm ci
  BUILD_CMD := npm run build
  TEST_CMD := npm test
  LINT_CMD := npm run lint
  DEV_CMD := npm run dev
  CLEAN_CMD := npm run clean
endif

# Node targets
.PHONY: install
install: ## Install dependencies
	$(call log_info,"Installing dependencies...")
	@$(INSTALL_CMD)
	$(call log_success,"Dependencies installed!")

.PHONY: build
build: ## Build Node.js project
	$(call log_info,"Building $(SERVICE_NAME)...")
	@$(BUILD_CMD)
	$(call log_success,"Build complete!")

.PHONY: test
test: ## Run Node.js tests
	$(call log_info,"Running tests...")
	@$(TEST_CMD)
	$(call log_success,"Tests complete!")

.PHONY: lint
lint: ## Run Node.js linters
	$(call log_info,"Running linters...")
	@$(LINT_CMD)
	$(call log_success,"Linting complete!")

.PHONY: dev
dev: ## Start development server
	$(call log_info,"Starting development server...")
	@$(DEV_CMD)

.PHONY: clean
clean: ## Clean build artifacts
	$(call log_info,"Cleaning build artifacts...")
	@rm -rf dist/ build/ .next/ out/
	@$(CLEAN_CMD) || true
	$(call log_success,"Clean complete!")
EOF

        # docker.mk
        cat > "$REPO_ROOT/build/makefiles/docker.mk" << 'EOF'
# Phoenix Platform - Docker Makefile

include $(ROOT_DIR)/build/makefiles/common.mk

# Docker variables
DOCKER_REGISTRY ?= ghcr.io/phoenix
DOCKER_IMAGE ?= $(DOCKER_REGISTRY)/$(SERVICE_NAME)
DOCKER_TAG ?= $(VERSION)
FULL_IMAGE_NAME = $(DOCKER_IMAGE):$(DOCKER_TAG)

# Docker targets
.PHONY: docker-build
docker-build: ## Build Docker image
	$(call log_info,"Building Docker image $(FULL_IMAGE_NAME)...")
	@docker build -t $(FULL_IMAGE_NAME) \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		.
	$(call log_success,"Docker image built!")

.PHONY: docker-push
docker-push: docker-build ## Push Docker image
	$(call log_info,"Pushing Docker image $(FULL_IMAGE_NAME)...")
	@docker push $(FULL_IMAGE_NAME)
	$(call log_success,"Docker image pushed!")

.PHONY: docker-run
docker-run: docker-build ## Run Docker container locally
	$(call log_info,"Running Docker container...")
	@docker run --rm -it $(DOCKER_PORT_ARGS) $(FULL_IMAGE_NAME)
EOF

        echo -e "${GREEN}✓${NC} Standard Makefiles created"
    else
        echo -e "${GREEN}✓${NC} Standard Makefiles already exist"
    fi
}

function setup_one_command_env() {
    echo -e "\n${YELLOW}Setting up one-command developer environment script...${NC}"

    # Create a one-command developer environment setup script
    cat > "$REPO_ROOT/scripts/setup-onecmd.sh" << 'EOF'
#!/bin/bash
# setup-onecmd.sh - One-command developer environment setup

set -euo pipefail

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo -e "${BLUE}=== Phoenix Platform - One Command Setup ===${NC}"
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}ERROR: Docker is not running. Please start Docker first.${NC}"
    exit 1
fi

# Step 1: Set up the development environment
echo -e "${YELLOW}1. Setting up development environment...${NC}"
"$REPO_ROOT/scripts/standardize-dev-env.sh" || {
    echo -e "${RED}Failed to set up development environment${NC}"
    exit 1
}

# Step 2: Create environment file
echo -e "\n${YELLOW}2. Setting up environment variables...${NC}"
"$REPO_ROOT/scripts/setup-dev-env.sh" || {
    echo -e "${RED}Failed to set up environment variables${NC}"
    exit 1
}

# Step 3: Start development dependencies
echo -e "\n${YELLOW}3. Starting development dependencies...${NC}"
make -f "$REPO_ROOT/Makefile" dev-up || {
    echo -e "${RED}Failed to start development dependencies${NC}"
    exit 1
}

# Step 4: Initialize the database if needed
if [ -f "$REPO_ROOT/Makefile.dev" ]; then
    echo -e "\n${YELLOW}4. Setting up database...${NC}"
    make -f "$REPO_ROOT/Makefile.dev" migrate-up || {
        echo -e "${YELLOW}Warning: Migration failed, but continuing...${NC}"
    }
fi

# Step 5: Install project dependencies
echo -e "\n${YELLOW}5. Installing project dependencies...${NC}"
for project in "$REPO_ROOT/projects"/*; do
    if [ -d "$project" ]; then
        project_name=$(basename "$project")
        echo -e "   Setting up project: ${BLUE}$project_name${NC}"
        
        # Go projects
        if [ -f "$project/go.mod" ]; then
            echo "   - Running go mod tidy"
            (cd "$project" && go mod tidy) || echo -e "${YELLOW}Warning: go mod tidy failed for $project_name${NC}"
        fi
        
        # Node.js projects
        if [ -f "$project/package.json" ]; then
            echo "   - Installing Node.js dependencies"
            if [ -f "$project/pnpm-lock.yaml" ]; then
                (cd "$project" && pnpm install) || echo -e "${YELLOW}Warning: pnpm install failed for $project_name${NC}"
            elif [ -f "$project/yarn.lock" ]; then
                (cd "$project" && yarn install) || echo -e "${YELLOW}Warning: yarn install failed for $project_name${NC}"
            else
                (cd "$project" && npm install) || echo -e "${YELLOW}Warning: npm install failed for $project_name${NC}"
            fi
        fi
    fi
done

echo -e "\n${GREEN}=== Setup Complete! ===${NC}"
echo ""
echo "Your development environment is ready. To start working:"
echo ""
echo "1. Start services:"
echo "   - All services: make dev"
echo "   - Specific service: cd projects/<service-name> && make run"
echo ""
echo "2. Access tools:"
echo "   - API: http://localhost:8080"
echo "   - Dashboard: http://localhost:3001"
echo "   - Prometheus: http://localhost:9090"
echo "   - Grafana: http://localhost:3000"
echo ""
echo "3. Common commands:"
echo "   - Build all: make build"
echo "   - Test all: make test"
echo "   - Lint all: make lint"
echo "   - Format code: make fmt"
echo ""
echo "4. To stop all services: make dev-down"
echo ""
echo "Happy coding! 🚀"
EOF

    chmod +x "$REPO_ROOT/scripts/setup-onecmd.sh"
    echo -e "${GREEN}✓${NC} One-command environment setup script created: scripts/setup-onecmd.sh"
}

function display_final_message() {
    echo -e "\n${GREEN}=== Phoenix Platform Developer Environment Standardization Complete ===${NC}"
    echo ""
    echo "Your development environment has been standardized. To get started:"
    echo ""
    echo "1. One-command setup for new team members:"
    echo "   ./scripts/setup-onecmd.sh"
    echo ""
    echo "2. Manual setup components:"
    echo "   - Environment configuration: ./scripts/setup-dev-env.sh"
    echo "   - Start services: make dev-up"
    echo ""
    echo "The following have been standardized:"
    echo "✓ Development tool versions"
    echo "✓ VS Code configuration"
    echo "✓ Git hooks and commit standards"
    echo "✓ Build system with common Makefiles"
    echo "✓ Environment variables"
    echo ""
}

# Main execution
prereqs_ok=true
check_prerequisites || prereqs_ok=false

if [ "$prereqs_ok" = false ]; then
    echo -e "\n${YELLOW}Some prerequisites are missing or outdated.${NC}"
    echo "Do you want to continue anyway? (y/n)"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        echo "Exiting. Please install the required prerequisites and try again."
        exit 1
    fi
fi

install_required_tools
configure_environments
setup_one_command_env
display_final_message
