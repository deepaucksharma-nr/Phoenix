-- Drop triggers
DROP TRIGGER IF EXISTS update_active_pipelines_updated_at ON active_pipelines;
DROP TRIGGER IF EXISTS update_tasks_updated_at ON tasks;
DROP TRIGGER IF EXISTS update_agents_updated_at ON agents;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_metric_cache_timestamp;
DROP INDEX IF EXISTS idx_metric_cache_experiment;
DROP INDEX IF EXISTS idx_active_pipelines_experiment;
DROP INDEX IF EXISTS idx_active_pipelines_host;
DROP INDEX IF EXISTS idx_experiment_events_experiment;
DROP INDEX IF EXISTS idx_agents_heartbeat;
DROP INDEX IF EXISTS idx_agents_status;
DROP INDEX IF EXISTS idx_tasks_experiment;
DROP INDEX IF EXISTS idx_tasks_host_status;

-- Drop tables
DROP TABLE IF EXISTS metric_cache;
DROP TABLE IF EXISTS active_pipelines;
DROP TABLE IF EXISTS experiment_events;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS agents;