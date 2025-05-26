-- Initial schema for Phoenix Platform
-- Creates the base tables for experiments, pipelines, and metrics

-- Create experiments table
CREATE TABLE IF NOT EXISTS experiments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    baseline_pipeline VARCHAR(255) NOT NULL,
    candidate_pipeline VARCHAR(255) NOT NULL,
    target_nodes JSONB,
    metrics JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    created_by VARCHAR(255),
    CONSTRAINT experiments_name_unique UNIQUE (name)
);

-- Create index on status for faster queries
CREATE INDEX idx_experiments_status ON experiments(status);
CREATE INDEX idx_experiments_created_at ON experiments(created_at DESC);

-- Create experiment_states table for state history
CREATE TABLE IF NOT EXISTS experiment_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    from_state VARCHAR(50),
    to_state VARCHAR(50) NOT NULL,
    reason TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255)
);

-- Create index for state history queries
CREATE INDEX idx_experiment_states_experiment_id ON experiment_states(experiment_id);
CREATE INDEX idx_experiment_states_created_at ON experiment_states(created_at DESC);

-- Create pipeline_deployments table
CREATE TABLE IF NOT EXISTS pipeline_deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    namespace VARCHAR(255) NOT NULL DEFAULT 'default',
    pipeline_name VARCHAR(255) NOT NULL,
    pipeline_config JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    replicas INTEGER DEFAULT 1,
    resources JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deployed_at TIMESTAMP WITH TIME ZONE,
    created_by VARCHAR(255),
    CONSTRAINT pipeline_deployments_unique UNIQUE (name, namespace)
);

-- Create indexes for pipeline deployments
CREATE INDEX idx_pipeline_deployments_namespace ON pipeline_deployments(namespace);
CREATE INDEX idx_pipeline_deployments_status ON pipeline_deployments(status);
CREATE INDEX idx_pipeline_deployments_pipeline_name ON pipeline_deployments(pipeline_name);

-- Create metrics table for experiment metrics
CREATE TABLE IF NOT EXISTS experiment_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    pipeline_type VARCHAR(50) NOT NULL, -- 'baseline' or 'candidate'
    metric_name VARCHAR(255) NOT NULL,
    metric_value DOUBLE PRECISION NOT NULL,
    labels JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for metrics queries
CREATE INDEX idx_experiment_metrics_experiment_id ON experiment_metrics(experiment_id);
CREATE INDEX idx_experiment_metrics_timestamp ON experiment_metrics(timestamp DESC);
CREATE INDEX idx_experiment_metrics_pipeline_type ON experiment_metrics(pipeline_type);
CREATE INDEX idx_experiment_metrics_metric_name ON experiment_metrics(metric_name);

-- Create hypertable for time-series metrics if TimescaleDB is available
-- SELECT create_hypertable('experiment_metrics', 'timestamp', if_not_exists => TRUE);

-- Create experiment_results table
CREATE TABLE IF NOT EXISTS experiment_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    cost_reduction_percent DOUBLE PRECISION,
    cardinality_reduction_percent DOUBLE PRECISION,
    performance_impact_percent DOUBLE PRECISION,
    monthly_savings_usd DOUBLE PRECISION,
    recommendation VARCHAR(50), -- 'promote', 'reject', 'extend'
    confidence_score DOUBLE PRECISION,
    analysis_details JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT experiment_results_unique UNIQUE (experiment_id)
);

-- Create audit_log table for tracking changes
CREATE TABLE IF NOT EXISTS audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    changes JSONB,
    user_id VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for audit log
CREATE INDEX idx_audit_log_entity ON audit_log(entity_type, entity_id);
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at DESC);
CREATE INDEX idx_audit_log_user_id ON audit_log(user_id);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_experiments_updated_at BEFORE UPDATE ON experiments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_pipeline_deployments_updated_at BEFORE UPDATE ON pipeline_deployments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create view for experiment summary
CREATE OR REPLACE VIEW experiment_summary AS
SELECT 
    e.id,
    e.name,
    e.description,
    e.status,
    e.baseline_pipeline,
    e.candidate_pipeline,
    e.created_at,
    e.started_at,
    e.ended_at,
    er.cost_reduction_percent,
    er.cardinality_reduction_percent,
    er.monthly_savings_usd,
    er.recommendation,
    er.confidence_score,
    COUNT(DISTINCT em.id) as metric_count,
    COUNT(DISTINCT es.id) as state_change_count
FROM experiments e
LEFT JOIN experiment_results er ON e.id = er.experiment_id
LEFT JOIN experiment_metrics em ON e.id = em.experiment_id
LEFT JOIN experiment_states es ON e.id = es.experiment_id
GROUP BY e.id, e.name, e.description, e.status, e.baseline_pipeline, 
         e.candidate_pipeline, e.created_at, e.started_at, e.ended_at,
         er.cost_reduction_percent, er.cardinality_reduction_percent,
         er.monthly_savings_usd, er.recommendation, er.confidence_score;

-- Sample data for development
INSERT INTO experiments (name, description, status, baseline_pipeline, candidate_pipeline, target_nodes) VALUES
('prometheus-optimization-v1', 'Optimize Prometheus metric collection', 'running', 'prometheus-baseline', 'prometheus-optimized', '{"prometheus": "prometheus-0"}'),
('datadog-tag-reduction', 'Reduce Datadog tag cardinality', 'completed', 'datadog-baseline', 'datadog-filtered', '{"datadog-agent": "datadog-agent-0"}'),
('newrelic-sampling', 'Implement intelligent sampling for New Relic', 'pending', 'newrelic-baseline', 'newrelic-sampled', '{"newrelic-infra": "newrelic-infra-0"}')
ON CONFLICT (name) DO NOTHING;

-- Add sample results for completed experiment
INSERT INTO experiment_results (experiment_id, cost_reduction_percent, cardinality_reduction_percent, performance_impact_percent, monthly_savings_usd, recommendation, confidence_score)
SELECT id, 72.8, 85.3, 0.5, 45000, 'promote', 0.95
FROM experiments WHERE name = 'datadog-tag-reduction'
ON CONFLICT (experiment_id) DO NOTHING;