-- Drop function
DROP FUNCTION IF EXISTS refresh_cardinality_summary();

-- Drop materialized view
DROP MATERIALIZED VIEW IF EXISTS cardinality_summary;

-- Drop tables
DROP TABLE IF EXISTS cardinality_analysis;
DROP TABLE IF EXISTS metrics;