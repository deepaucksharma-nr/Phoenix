-- Remove NRDOT support

-- Drop collector metrics table
DROP TABLE IF EXISTS collector_metrics;

-- Remove indexes
DROP INDEX IF EXISTS idx_pipeline_deployments_collector_config;
DROP INDEX IF EXISTS idx_experiments_collector_type;

-- Remove columns
ALTER TABLE pipeline_deployments DROP COLUMN IF EXISTS collector_config;
ALTER TABLE experiments DROP COLUMN IF EXISTS collector_type;