-- Phoenix Lean-Core Architecture Database Schema
-- This migration adds tables needed for the agent-based architecture

BEGIN;

-- Agent task queue for distributing work to agents
CREATE TABLE IF NOT EXISTS agent_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    host_id VARCHAR(255) NOT NULL,
    experiment_id VARCHAR(50) REFERENCES experiments(id) ON DELETE CASCADE,
    task_type VARCHAR(50) NOT NULL CHECK (task_type IN ('collector', 'loadsim', 'command')),
    action VARCHAR(50) NOT NULL CHECK (action IN ('start', 'stop', 'update', 'execute')),
    config JSONB NOT NULL,
    priority INT DEFAULT 0,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'assigned', 'running', 'completed', 'failed')),
    assigned_at TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    result JSONB,
    error_message TEXT,
    retry_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for efficient task polling
CREATE INDEX idx_agent_tasks_host_status ON agent_tasks(host_id, status);
CREATE INDEX idx_agent_tasks_experiment ON agent_tasks(experiment_id);
CREATE INDEX idx_agent_tasks_created ON agent_tasks(created_at);
CREATE INDEX idx_agent_tasks_priority ON agent_tasks(priority DESC, created_at ASC);

-- Agent heartbeat and status tracking
CREATE TABLE IF NOT EXISTS agent_status (
    host_id VARCHAR(255) PRIMARY KEY,
    hostname VARCHAR(255),
    ip_address INET,
    agent_version VARCHAR(50),
    started_at TIMESTAMP,
    last_heartbeat TIMESTAMP NOT NULL,
    status VARCHAR(50) DEFAULT 'healthy' CHECK (status IN ('healthy', 'degraded', 'unhealthy', 'offline')),
    capabilities JSONB DEFAULT '{}',
    active_tasks JSONB DEFAULT '[]',
    resource_usage JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_agent_status_heartbeat ON agent_status(last_heartbeat);

-- Active pipelines tracking (replaces K8s CRD state)
CREATE TABLE IF NOT EXISTS active_pipelines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    host_id VARCHAR(255) NOT NULL,
    experiment_id VARCHAR(50) REFERENCES experiments(id) ON DELETE CASCADE,
    variant VARCHAR(50) NOT NULL CHECK (variant IN ('baseline', 'candidate')),
    config_url TEXT NOT NULL,
    config_hash VARCHAR(64),
    process_info JSONB DEFAULT '{}',
    metrics_info JSONB DEFAULT '{}',
    status VARCHAR(50) DEFAULT 'running' CHECK (status IN ('starting', 'running', 'stopping', 'stopped', 'failed')),
    started_at TIMESTAMP DEFAULT NOW(),
    stopped_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(host_id, experiment_id, variant)
);

CREATE INDEX idx_active_pipelines_experiment ON active_pipelines(experiment_id);
CREATE INDEX idx_active_pipelines_host ON active_pipelines(host_id);

-- Metrics cache for faster queries
CREATE TABLE IF NOT EXISTS metrics_cache (
    id SERIAL PRIMARY KEY,
    experiment_id VARCHAR(50) REFERENCES experiments(id) ON DELETE CASCADE,
    timestamp TIMESTAMP NOT NULL,
    metric_name VARCHAR(255) NOT NULL,
    variant VARCHAR(50) NOT NULL,
    host_id VARCHAR(255),
    value DOUBLE PRECISION,
    labels JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_metrics_cache_experiment_time ON metrics_cache(experiment_id, timestamp);
CREATE INDEX idx_metrics_cache_metric_variant ON metrics_cache(metric_name, variant);
CREATE INDEX idx_metrics_cache_timestamp ON metrics_cache(timestamp);

-- Add columns to existing experiments table for lean-core support
ALTER TABLE experiments ADD COLUMN IF NOT EXISTS deployment_mode VARCHAR(50) DEFAULT 'kubernetes';
ALTER TABLE experiments ADD COLUMN IF NOT EXISTS target_hosts TEXT[];

-- Create views for backward compatibility
CREATE OR REPLACE VIEW deployment_status AS
SELECT 
    e.id as experiment_id,
    e.phase,
    COUNT(DISTINCT ap.host_id) as active_hosts,
    jsonb_agg(jsonb_build_object(
        'host', ap.host_id,
        'baseline_status', MAX(CASE WHEN ap.variant = 'baseline' THEN ap.status END),
        'candidate_status', MAX(CASE WHEN ap.variant = 'candidate' THEN ap.status END)
    )) as host_details
FROM experiments e
LEFT JOIN active_pipelines ap ON e.id = ap.experiment_id
GROUP BY e.id, e.phase;

-- Function to update timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Add update triggers
CREATE TRIGGER update_agent_tasks_updated_at BEFORE UPDATE ON agent_tasks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_agent_status_updated_at BEFORE UPDATE ON agent_status
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_active_pipelines_updated_at BEFORE UPDATE ON active_pipelines
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMIT;