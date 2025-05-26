#!/bin/bash

# Fix Phoenix CLI command naming consistency
cd /Users/deepaksharma/Desktop/src/Phoenix/projects/phoenix-cli/cmd

echo "Fixing CLI command variable names..."

# Fix remaining pipeline commands
sed -i '' 's/var deployPipelineCmd/var pipelineDeployCmd/g' pipeline_deploy.go
sed -i '' 's/deployPipelineCmd\.AddCommand/pipelineDeployCmd.AddCommand/g' pipeline_deploy.go
sed -i '' 's/deployPipelineCmd\.Flags/pipelineDeployCmd.Flags/g' pipeline_deploy.go

# Fix list deployments command
sed -i '' 's/var listDeploymentsCmd/var pipelineListDeploymentsCmd/g' pipeline_list_deployments.go
sed -i '' 's/listDeploymentsCmd\.AddCommand/pipelineListDeploymentsCmd.AddCommand/g' pipeline_list_deployments.go
sed -i '' 's/listDeploymentsCmd\.Flags/pipelineListDeploymentsCmd.Flags/g' pipeline_list_deployments.go

# Fix show command
sed -i '' 's/var showPipelineCmd/var pipelineShowCmd/g' pipeline_show.go
sed -i '' 's/showPipelineCmd\.AddCommand/pipelineShowCmd.AddCommand/g' pipeline_show.go
sed -i '' 's/showPipelineCmd\.Flags/pipelineShowCmd.Flags/g' pipeline_show.go

# Fix validate command
sed -i '' 's/var validatePipelineCmd/var pipelineValidateCmd/g' pipeline_validate.go
sed -i '' 's/validatePipelineCmd\.AddCommand/pipelineValidateCmd.AddCommand/g' pipeline_validate.go
sed -i '' 's/validatePipelineCmd\.Flags/pipelineValidateCmd.Flags/g' pipeline_validate.go

# Fix status command
sed -i '' 's/var statusPipelineCmd/var pipelineStatusCmd/g' pipeline_status.go
sed -i '' 's/statusPipelineCmd\.AddCommand/pipelineStatusCmd.AddCommand/g' pipeline_status.go
sed -i '' 's/statusPipelineCmd\.Flags/pipelineStatusCmd.Flags/g' pipeline_status.go

# Fix get-config command
sed -i '' 's/var getConfigCmd/var pipelineGetConfigCmd/g' pipeline_get_config.go
sed -i '' 's/getConfigCmd\.AddCommand/pipelineGetConfigCmd.AddCommand/g' pipeline_get_config.go
sed -i '' 's/getConfigCmd\.Flags/pipelineGetConfigCmd.Flags/g' pipeline_get_config.go

# Fix rollback command
sed -i '' 's/var rollbackPipelineCmd/var pipelineRollbackCmd/g' pipeline_rollback.go
sed -i '' 's/rollbackPipelineCmd\.AddCommand/pipelineRollbackCmd.AddCommand/g' pipeline_rollback.go
sed -i '' 's/rollbackPipelineCmd\.Flags/pipelineRollbackCmd.Flags/g' pipeline_rollback.go

# Fix delete command
sed -i '' 's/var deletePipelineCmd/var pipelineDeleteCmd/g' pipeline_delete.go
sed -i '' 's/deletePipelineCmd\.AddCommand/pipelineDeleteCmd.AddCommand/g' pipeline_delete.go
sed -i '' 's/deletePipelineCmd\.Flags/pipelineDeleteCmd.Flags/g' pipeline_delete.go

echo "✅ Fixed all CLI command variable names"

# Also fix the init functions to use correct variable names
for file in pipeline_*.go; do
    # Fix init functions that add to parent command
    sed -i '' "s/pipelineCmd\.AddCommand(\([^)]*\)Cmd)/pipelineCmd.AddCommand(\1Cmd)/g" "$file"
done

echo "✅ Fixed all init function references"