-- Rollback lean-core architecture changes

BEGIN;

-- Drop triggers
DROP TRIGGER IF EXISTS update_agent_tasks_updated_at ON agent_tasks;
DROP TRIGGER IF EXISTS update_agent_status_updated_at ON agent_status;
DROP TRIGGER IF EXISTS update_active_pipelines_updated_at ON active_pipelines;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop views
DROP VIEW IF EXISTS deployment_status;

-- Remove added columns from experiments
ALTER TABLE experiments DROP COLUMN IF EXISTS deployment_mode;
ALTER TABLE experiments DROP COLUMN IF EXISTS target_hosts;

-- Drop tables
DROP TABLE IF EXISTS metrics_cache;
DROP TABLE IF EXISTS active_pipelines;
DROP TABLE IF EXISTS agent_status;
DROP TABLE IF EXISTS agent_tasks;

COMMIT;