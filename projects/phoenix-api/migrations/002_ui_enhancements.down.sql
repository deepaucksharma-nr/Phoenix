-- Rollback UI Enhancement Tables

-- Drop triggers
DROP TRIGGER IF EXISTS update_agent_ui_state_updated_at ON agent_ui_state;
DROP TRIGGER IF EXISTS update_pipeline_templates_updated_at ON pipeline_templates;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse order
DROP TABLE IF EXISTS ui_activity_log;
DROP TABLE IF EXISTS experiment_wizard_history;
DROP TABLE IF EXISTS metric_flow_snapshots;
DROP TABLE IF EXISTS cost_analytics;
DROP TABLE IF EXISTS pipeline_templates;
DROP TABLE IF EXISTS agent_ui_state;
DROP TABLE IF EXISTS metric_cost_cache;