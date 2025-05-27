-- Phoenix Platform Initial Schema
-- Core tables required by the platform

BEGIN;

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Experiments table
CREATE TABLE IF NOT EXISTS experiments (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'created',
    phase VARCHAR(50) DEFAULT 'created',
    baseline_pipeline VARCHAR(255),
    candidate_pipeline VARCHAR(255),
    target_nodes TEXT[],
    config JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Pipeline deployments table
CREATE TABLE IF NOT EXISTS pipeline_deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    namespace VARCHAR(255) DEFAULT 'default',
    pipeline_config TEXT NOT NULL,
    variant VARCHAR(50) DEFAULT 'standalone',
    status VARCHAR(50) DEFAULT 'pending',
    status_message TEXT,
    parameters JSONB DEFAULT '{}',
    metrics JSONB DEFAULT '{}',
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Tasks table for agent communication
CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(50) NOT NULL,
    host_id VARCHAR(255),
    target_id VARCHAR(255),
    experiment_id VARCHAR(50),
    deployment_id UUID,
    action VARCHAR(50) NOT NULL,
    parameters JSONB DEFAULT '{}',
    status VARCHAR(50) DEFAULT 'pending',
    priority INT DEFAULT 0,
    assigned_at TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    result JSONB,
    error TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Agents table
CREATE TABLE IF NOT EXISTS agents (
    host_id VARCHAR(255) PRIMARY KEY,
    hostname VARCHAR(255),
    ip_address VARCHAR(45),
    version VARCHAR(50),
    status VARCHAR(50) DEFAULT 'healthy',
    capabilities JSONB DEFAULT '{}',
    metrics JSONB DEFAULT '{}',
    last_heartbeat TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Experiment events for tracking state changes
CREATE TABLE IF NOT EXISTS experiment_events (
    id SERIAL PRIMARY KEY,
    experiment_id VARCHAR(50) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    phase VARCHAR(50),
    message TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_experiments_status ON experiments(status);
CREATE INDEX idx_experiments_created ON experiments(created_at);
CREATE INDEX idx_pipeline_deployments_status ON pipeline_deployments(status);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_host ON tasks(host_id);
CREATE INDEX idx_tasks_experiment ON tasks(experiment_id);
CREATE INDEX idx_agents_heartbeat ON agents(last_heartbeat);
CREATE INDEX idx_experiment_events_experiment ON experiment_events(experiment_id);

COMMIT;