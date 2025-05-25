-- 001_create_experiments.sql
-- Core experiments table

-- Create enum for experiment status
CREATE TYPE experiment_status AS ENUM (
    'draft',
    'pending',
    'initializing',
    'deploying',
    'running',
    'analyzing',
    'completed',
    'failed',
    'cancelled'
);

-- Create enum for experiment type
CREATE TYPE experiment_type AS ENUM (
    'ab_test',
    'canary',
    'blue_green'
);

-- Main experiments table
CREATE TABLE IF NOT EXISTS experiments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type experiment_type NOT NULL DEFAULT 'ab_test',
    status experiment_status NOT NULL DEFAULT 'draft',
    
    -- Experiment configuration
    baseline_pipeline_id UUID,
    candidate_pipeline_id UUID,
    
    -- Targeting
    target_hosts JSONB NOT NULL DEFAULT '[]',
    target_labels JSONB NOT NULL DEFAULT '{}',
    
    -- Success criteria
    success_criteria JSONB NOT NULL DEFAULT '{}',
    
    -- Scheduling
    scheduled_start_time TIMESTAMP WITH TIME ZONE,
    scheduled_end_time TIMESTAMP WITH TIME ZONE,
    actual_start_time TIMESTAMP WITH TIME ZONE,
    actual_end_time TIMESTAMP WITH TIME ZONE,
    
    -- Results
    results JSONB,
    winner VARCHAR(50), -- 'baseline' or 'candidate'
    
    -- Metadata
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Soft delete
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT valid_times CHECK (
        (scheduled_end_time IS NULL OR scheduled_start_time IS NULL) OR 
        scheduled_end_time > scheduled_start_time
    ),
    CONSTRAINT valid_winner CHECK (winner IN ('baseline', 'candidate'))
);

-- Indexes for performance
CREATE INDEX idx_experiments_status ON experiments(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_experiments_type ON experiments(type) WHERE deleted_at IS NULL;
CREATE INDEX idx_experiments_created_at ON experiments(created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_experiments_scheduled_start ON experiments(scheduled_start_time) WHERE deleted_at IS NULL;
CREATE INDEX idx_experiments_created_by ON experiments(created_by) WHERE deleted_at IS NULL;

-- Trigger to update updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_experiments_updated_at BEFORE UPDATE
    ON experiments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Comments for documentation
COMMENT ON TABLE experiments IS 'Core table storing experiment configurations and results';
COMMENT ON COLUMN experiments.id IS 'Unique experiment identifier';
COMMENT ON COLUMN experiments.status IS 'Current state of the experiment in its lifecycle';
COMMENT ON COLUMN experiments.type IS 'Type of experiment (A/B test, canary, blue-green)';
COMMENT ON COLUMN experiments.target_hosts IS 'JSON array of host selectors';
COMMENT ON COLUMN experiments.target_labels IS 'JSON object of label selectors';
COMMENT ON COLUMN experiments.success_criteria IS 'JSON object defining success metrics';
COMMENT ON COLUMN experiments.results IS 'JSON object containing experiment results and metrics';