-- Drop functions
DROP FUNCTION IF EXISTS rollback_deployment_version(VARCHAR, INTEGER, VARCHAR, TEXT);
DROP FUNCTION IF EXISTS record_deployment_version(VARCHAR, TEXT, JSONB, VARCHAR, TEXT);
DROP FUNCTION IF EXISTS get_next_deployment_version(VARCHAR);

-- Drop columns from pipeline_deployments
ALTER TABLE pipeline_deployments 
DROP COLUMN IF EXISTS current_version,
DROP COLUMN IF EXISTS last_version_change;

-- Drop deployment versions table
DROP TABLE IF EXISTS deployment_versions;