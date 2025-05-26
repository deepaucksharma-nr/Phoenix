# dashboard Makefile
include ../../build/makefiles/common.mk
include ../../build/makefiles/node.mk
include ../../build/makefiles/docker.mk

PROJECT_NAME := dashboard

# Default target
.DEFAULT_GOAL := help

# Build targets
build: node-build
test: node-test
lint: node-lint
fmt: node-fmt
clean: node-clean

# Docker targets
docker: docker-build
docker-push: docker-push

# Development targets
dev: node-dev
preview: node-preview

# Type checking
typecheck: node-typecheck

.PHONY: all build test lint fmt clean docker docker-push dev preview typecheck
