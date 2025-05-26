-- Drop metrics-related tables and functions
DROP FUNCTION IF EXISTS clean_old_metrics(INTEGER);

DROP TABLE IF EXISTS metrics_buffer;
DROP TABLE IF EXISTS cardinality_analysis;
DROP TABLE IF EXISTS cost_tracking;
DROP TABLE IF EXISTS metrics_aggregated;
DROP TABLE IF EXISTS metrics;