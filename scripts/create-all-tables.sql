-- Create all necessary tables for Phoenix
BEGIN;

-- Create schema_migrations table if needed
CREATE TABLE IF NOT EXISTS schema_migrations (
    version bigint NOT NULL PRIMARY KEY,
    dirty boolean NOT NULL DEFAULT false
);

-- Create experiments table with all necessary columns
CREATE TABLE IF NOT EXISTS experiments (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    phase VARCHAR(50) DEFAULT 'pending',
    baseline_pipeline VARCHAR(255),
    candidate_pipeline VARCHAR(255),
    target_nodes JSONB DEFAULT '[]',
    config JSONB DEFAULT '{}',
    status JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    deployment_mode VARCHAR(50) DEFAULT 'kubernetes',
    target_hosts TEXT[],
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Agent task queue (renamed to tasks)
CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    host_id VARCHAR(255) NOT NULL,
    experiment_id VARCHAR(50),
    task_type VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    priority INT DEFAULT 0,
    status VARCHAR(50) DEFAULT 'pending',
    assigned_at TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    result JSONB,
    error_message TEXT,
    retry_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Agent status tracking (renamed to agents)
CREATE TABLE IF NOT EXISTS agents (
    host_id VARCHAR(255) PRIMARY KEY,
    hostname VARCHAR(255),
    ip_address INET,
    agent_version VARCHAR(50),
    started_at TIMESTAMP,
    last_heartbeat TIMESTAMP NOT NULL,
    status VARCHAR(50) DEFAULT 'healthy',
    capabilities JSONB DEFAULT '{}',
    active_tasks JSONB DEFAULT '[]',
    resource_usage JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create metrics cache
CREATE TABLE IF NOT EXISTS metric_cache (
    id SERIAL PRIMARY KEY,
    experiment_id VARCHAR(50),
    timestamp TIMESTAMP NOT NULL,
    metric_name VARCHAR(255) NOT NULL,
    variant VARCHAR(50) NOT NULL,
    host_id VARCHAR(255),
    value DOUBLE PRECISION,
    labels JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create experiment events table
CREATE TABLE IF NOT EXISTS experiment_events (
    id SERIAL PRIMARY KEY,
    experiment_id VARCHAR(50),
    event_type VARCHAR(100) NOT NULL,
    phase VARCHAR(50),
    message TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create pipeline deployments table
CREATE TABLE IF NOT EXISTS pipeline_deployments (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    namespace VARCHAR(255) DEFAULT 'default',
    pipeline_id VARCHAR(255),
    variant VARCHAR(50),
    status VARCHAR(50) DEFAULT 'pending',
    created_by VARCHAR(255),
    metrics JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create users table for auth
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_tasks_host_status ON tasks(host_id, status);
CREATE INDEX IF NOT EXISTS idx_tasks_experiment ON tasks(experiment_id);
CREATE INDEX IF NOT EXISTS idx_agents_heartbeat ON agents(last_heartbeat);
CREATE INDEX IF NOT EXISTS idx_metric_cache_experiment_time ON metric_cache(experiment_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_experiment_events_experiment ON experiment_events(experiment_id);

-- Insert schema version to prevent migrations
INSERT INTO schema_migrations (version, dirty) VALUES (5, false);

COMMIT;