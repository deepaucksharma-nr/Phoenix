-- Create agents table for agent status tracking
CREATE TABLE IF NOT EXISTS agents (
    host_id VARCHAR(255) PRIMARY KEY,
    hostname VARCHAR(255) NOT NULL,
    ip_address VARCHAR(50),
    agent_version VARCHAR(50),
    started_at TIMESTAMP,
    last_heartbeat TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) NOT NULL DEFAULT 'unknown',
    capabilities JSONB DEFAULT '{}',
    active_tasks TEXT[] DEFAULT '{}',
    resource_usage JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create tasks table for task queue
CREATE TABLE IF NOT EXISTS tasks (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid(),
    host_id VARCHAR(255) NOT NULL,
    experiment_id VARCHAR(255),
    task_type VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    config JSONB DEFAULT '{}',
    priority INT NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    assigned_at TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    result JSONB,
    error_message TEXT,
    retry_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_experiment
        FOREIGN KEY(experiment_id) 
        REFERENCES experiments(id)
        ON DELETE CASCADE
);

-- Create experiment_events table
CREATE TABLE IF NOT EXISTS experiment_events (
    id SERIAL PRIMARY KEY,
    experiment_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    phase VARCHAR(50),
    message TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_experiment_event
        FOREIGN KEY(experiment_id) 
        REFERENCES experiments(id)
        ON DELETE CASCADE
);

-- Create active_pipelines table to track running collectors
CREATE TABLE IF NOT EXISTS active_pipelines (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid(),
    host_id VARCHAR(255) NOT NULL,
    experiment_id VARCHAR(255) NOT NULL,
    variant VARCHAR(50) NOT NULL,
    config_url VARCHAR(1024),
    config_hash VARCHAR(64),
    process_info JSONB DEFAULT '{}',
    metrics_info JSONB DEFAULT '{}',
    status VARCHAR(50) NOT NULL DEFAULT 'starting',
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    stopped_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_experiment_pipeline
        FOREIGN KEY(experiment_id) 
        REFERENCES experiments(id)
        ON DELETE CASCADE
);

-- Create metric_cache table for fast queries
CREATE TABLE IF NOT EXISTS metric_cache (
    id SERIAL PRIMARY KEY,
    experiment_id VARCHAR(255) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    metric_name VARCHAR(255) NOT NULL,
    variant VARCHAR(50) NOT NULL,
    host_id VARCHAR(255) NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    labels JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_experiment_metric
        FOREIGN KEY(experiment_id) 
        REFERENCES experiments(id)
        ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX idx_tasks_host_status ON tasks(host_id, status);
CREATE INDEX idx_tasks_experiment ON tasks(experiment_id);
CREATE INDEX idx_agents_status ON agents(status);
CREATE INDEX idx_agents_heartbeat ON agents(last_heartbeat);
CREATE INDEX idx_experiment_events_experiment ON experiment_events(experiment_id);
CREATE INDEX idx_active_pipelines_host ON active_pipelines(host_id);
CREATE INDEX idx_active_pipelines_experiment ON active_pipelines(experiment_id);
CREATE INDEX idx_metric_cache_experiment ON metric_cache(experiment_id, variant);
CREATE INDEX idx_metric_cache_timestamp ON metric_cache(timestamp);

-- Add triggers to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_agents_updated_at BEFORE UPDATE ON agents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tasks_updated_at BEFORE UPDATE ON tasks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_active_pipelines_updated_at BEFORE UPDATE ON active_pipelines
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();