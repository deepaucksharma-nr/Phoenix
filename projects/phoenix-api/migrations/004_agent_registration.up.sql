-- Agent registration and health tracking
CREATE TABLE IF NOT EXISTS agents (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid(),
    host_id VARCHAR(255) NOT NULL UNIQUE,
    hostname VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45),
    platform VARCHAR(50),
    os_version VARCHAR(100),
    agent_version VARCHAR(50),
    status VARCHAR(50) NOT NULL DEFAULT 'online',
    last_heartbeat TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    registered_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB DEFAULT '{}'
);

-- Index for quick status queries
CREATE INDEX idx_agents_status ON agents(status);
CREATE INDEX idx_agents_last_heartbeat ON agents(last_heartbeat);

-- Agent capabilities tracking
CREATE TABLE IF NOT EXISTS agent_capabilities (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id VARCHAR(255) NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    capability VARCHAR(100) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    config JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(agent_id, capability)
);

-- Agent resource usage tracking
CREATE TABLE IF NOT EXISTS agent_resources (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id VARCHAR(255) NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    cpu_percent FLOAT,
    memory_percent FLOAT,
    memory_used_mb BIGINT,
    memory_total_mb BIGINT,
    disk_percent FLOAT,
    disk_used_gb BIGINT,
    disk_total_gb BIGINT,
    network_rx_mbps FLOAT,
    network_tx_mbps FLOAT,
    load_average_1m FLOAT,
    load_average_5m FLOAT,
    load_average_15m FLOAT,
    process_count INT,
    goroutine_count INT
);

-- Index for time-series queries
CREATE INDEX idx_agent_resources_agent_timestamp ON agent_resources(agent_id, timestamp DESC);

-- Create hypertable for time-series data if TimescaleDB is available
-- COMMENTED OUT: Uncomment if using TimescaleDB
-- SELECT create_hypertable('agent_resources', 'timestamp', if_not_exists => TRUE);

-- Update trigger for agents table
CREATE OR REPLACE FUNCTION update_agents_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER agents_updated_at_trigger
    BEFORE UPDATE ON agents
    FOR EACH ROW
    EXECUTE FUNCTION update_agents_updated_at();

-- Add agent_id to tasks table if not exists
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name='tasks' AND column_name='agent_id') THEN
        ALTER TABLE tasks ADD COLUMN agent_id VARCHAR(255);
        ALTER TABLE tasks ADD CONSTRAINT fk_tasks_agent 
            FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE SET NULL;
    END IF;
END $$;

-- Add index for agent tasks
CREATE INDEX IF NOT EXISTS idx_tasks_agent_id ON tasks(agent_id);