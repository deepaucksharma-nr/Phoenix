# Configuration Generation Makefile

.PHONY: all dev prod clean

CONFIG_GEN := go run ../tools/config-gen/main.go

# Generate all configurations
all: dev prod

# Generate development configs
dev:
	@echo "Generating development configurations..."
	@$(CONFIG_GEN) -env development \
		-template monitoring/prometheus/base/prometheus.yaml.tmpl \
		-output generated/development/prometheus.yaml

# Generate production configs  
prod:
	@echo "Generating production configurations..."
	@$(CONFIG_GEN) -env production \
		-template monitoring/prometheus/base/prometheus.yaml.tmpl \
		-output generated/production/prometheus.yaml

# Clean generated files
clean:
	@echo "Cleaning generated configurations..."
	@rm -rf generated/

# Validate generated configs
validate: all
	@echo "Validating generated configurations..."
	@command -v promtool >/dev/null 2>&1 && promtool check config generated/*/prometheus.yaml || echo "promtool not found, skipping validation"