-- Create experiments table with all necessary columns
CREATE TABLE IF NOT EXISTS experiments (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    phase VARCHAR(50) DEFAULT 'pending',
    config JSONB DEFAULT '{}',
    status JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    deployment_mode VARCHAR(50) DEFAULT 'kubernetes',
    target_hosts TEXT[],
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create experiment events table
CREATE TABLE IF NOT EXISTS experiment_events (
    id SERIAL PRIMARY KEY,
    experiment_id VARCHAR(50) REFERENCES experiments(id) ON DELETE CASCADE,
    event_type VARCHAR(100) NOT NULL,
    phase VARCHAR(50),
    message TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create metrics table
CREATE TABLE IF NOT EXISTS metrics (
    id SERIAL PRIMARY KEY,
    experiment_id VARCHAR(50),
    metric_name VARCHAR(255),
    value DOUBLE PRECISION,
    variant VARCHAR(50),
    timestamp TIMESTAMP DEFAULT NOW(),
    labels JSONB DEFAULT '{}'
);