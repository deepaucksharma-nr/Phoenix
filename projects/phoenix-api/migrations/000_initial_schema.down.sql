-- Drop Phoenix Platform initial schema

BEGIN;

DROP TABLE IF EXISTS experiment_events CASCADE;
DROP TABLE IF EXISTS agents CASCADE;
DROP TABLE IF EXISTS tasks CASCADE;
DROP TABLE IF EXISTS pipeline_deployments CASCADE;
DROP TABLE IF EXISTS experiments CASCADE;

COMMIT;