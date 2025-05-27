-- Drop views and functions
DROP VIEW IF EXISTS metric_cost_flow_view;
DROP FUNCTION IF EXISTS calculate_metric_cost;

-- Drop tables
DROP TABLE IF EXISTS pipeline_templates;
DROP TABLE IF EXISTS cost_tracking;
DROP TABLE IF EXISTS cardinality_analysis;
DROP TABLE IF EXISTS metric_cache;
DROP TABLE IF EXISTS metrics;