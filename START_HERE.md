# 🚀 Phoenix Platform - START HERE

Welcome to the newly migrated Phoenix Platform! This guide will get you up and running in minutes.

## 🎯 Quick Start (3 Steps)

### Step 1: Install Dependencies
```bash
# Install protoc (choose your OS)
# macOS:
brew install protobuf

# Ubuntu/Debian:
sudo apt-get install -y protobuf-compiler

# Install Go protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Step 2: Build Phoenix CLI
```bash
# Navigate to Phoenix CLI
cd projects/phoenix-cli

# Build the CLI
go build -o bin/phoenix .

# Add to PATH (optional)
export PATH=$PATH:$(pwd)/bin

# Verify it works
./bin/phoenix --help
```

### Step 3: Start Using Phoenix
```bash
# Create an experiment
phoenix experiment create --name "my-test" --baseline "baseline-v1" --candidate "optimized-v1"

# List experiments  
phoenix experiment list

# Check status
phoenix experiment status <id>
```

## 📁 Project Structure

```
Phoenix/
├── projects/phoenix-cli/    ← Phoenix CLI is here!
├── services/               ← Core services
├── packages/              ← Shared packages
├── operators/             ← K8s operators
└── infrastructure/        ← Deployment configs
```

## 📚 Documentation

- **[DEVELOPMENT_GUIDE.md](DEVELOPMENT_GUIDE.md)** - Full developer guide
- **[QUICK_START.md](QUICK_START.md)** - Detailed quick start
- **[MIGRATION_FINAL_SUCCESS.md](MIGRATION_FINAL_SUCCESS.md)** - Migration details

## 🔧 Common Commands

```bash
# Build all services
go work sync
cd services/api && go build ./cmd/main.go
cd ../controller && go build ./cmd/controller/main.go

# Run tests
go test ./...

# Generate protos
cd packages/contracts && bash generate.sh
```

## ❓ Need Help?

1. Check the [DEVELOPMENT_GUIDE.md](DEVELOPMENT_GUIDE.md)
2. Review example code in tests
3. Look at the Phoenix CLI help: `phoenix --help`

---

**🎉 Welcome to Phoenix Platform - Ready to optimize your observability costs!**