#!/bin/bash

# Build Phoenix CLI with only existing commands

cd "$(dirname "$0")"

# List of command files to build
FILES=(
    root.go
    version.go
    experiment.go
    experiment_create.go
    experiment_list.go
    experiment_start.go
    experiment_stop.go
    experiment_status.go
    experiment_metrics.go
    experiment_promote.go
    config.go
    completion.go
    auth.go
    auth_login.go
    auth_logout.go
    auth_status.go
    benchmark.go
    migrate.go
    plugin.go
    loadsim.go
    loadsim_start.go
    loadsim_stop.go
    loadsim_status.go
    loadsim_list_profiles.go
)

# Comment out pipeline commands for now
# pipeline.go
# pipeline_create.go
# pipeline_list.go
# pipeline_deploy.go
# pipeline_list_deployments.go

echo "Building Phoenix CLI..."
go build -o ../bin/phoenix-cli "${FILES[@]}"