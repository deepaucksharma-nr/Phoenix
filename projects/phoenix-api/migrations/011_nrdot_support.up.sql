-- Add NRDOT support to experiments and deployments

-- Add collector_type to experiments if not exists
ALTER TABLE experiments 
ADD COLUMN IF NOT EXISTS collector_type VARCHAR(50) DEFAULT 'otel';

-- Add NRDOT configuration to pipeline_deployments
ALTER TABLE pipeline_deployments
ADD COLUMN IF NOT EXISTS collector_config JSONB DEFAULT '{}';

-- Add index for collector type queries
CREATE INDEX IF NOT EXISTS idx_experiments_collector_type ON experiments(collector_type);
CREATE INDEX IF NOT EXISTS idx_pipeline_deployments_collector_config ON pipeline_deployments USING GIN(collector_config);

-- Update existing experiments to set collector_type based on pipeline
UPDATE experiments 
SET collector_type = 'nrdot'
WHERE metadata->>'candidate_pipeline' LIKE 'nrdot-%'
   OR metadata->>'baseline_pipeline' LIKE 'nrdot-%';

-- Add collector metrics table for tracking collector-specific metrics
CREATE TABLE IF NOT EXISTS collector_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    agent_id VARCHAR(255) NOT NULL,
    collector_type VARCHAR(50) NOT NULL,
    variant VARCHAR(50) NOT NULL,
    metrics JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Indexes
    INDEX idx_collector_metrics_experiment (experiment_id),
    INDEX idx_collector_metrics_agent (agent_id),
    INDEX idx_collector_metrics_type (collector_type),
    INDEX idx_collector_metrics_created (created_at DESC)
);

-- Add comment
COMMENT ON TABLE collector_metrics IS 'Stores collector-specific metrics including NRDOT cardinality reduction data';
COMMENT ON COLUMN collector_metrics.collector_type IS 'Type of collector: otel or nrdot';
COMMENT ON COLUMN collector_metrics.metrics IS 'Collector-specific metrics data';