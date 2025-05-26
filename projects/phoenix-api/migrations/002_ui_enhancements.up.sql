-- UI Enhancement Tables for Phoenix Platform

-- Metric cost cache for instant calculations
CREATE TABLE IF NOT EXISTS metric_cost_cache (
    metric_name TEXT PRIMARY KEY,
    cardinality BIGINT NOT NULL DEFAULT 0,
    cost_per_minute DECIMAL(10,2) NOT NULL DEFAULT 0,
    last_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    labels JSONB DEFAULT '{}',
    namespace TEXT,
    service TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX idx_metric_cost_cache_cost ON metric_cost_cache(cost_per_minute DESC);
CREATE INDEX idx_metric_cost_cache_namespace ON metric_cost_cache(namespace);
CREATE INDEX idx_metric_cost_cache_service ON metric_cost_cache(service);
CREATE INDEX idx_metric_cost_cache_updated ON metric_cost_cache(last_updated);

-- Agent UI state for enhanced visualization
CREATE TABLE IF NOT EXISTS agent_ui_state (
    host_id TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    group_name TEXT,
    location JSONB DEFAULT NULL, -- {"latitude": 0, "longitude": 0, "region": "", "zone": ""}
    ui_metadata JSONB DEFAULT '{}', -- Custom UI properties
    color TEXT, -- For visual grouping
    icon TEXT, -- Custom icon identifier
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_agent_ui_state_group ON agent_ui_state(group_name);

-- Pipeline templates for wizard
CREATE TABLE IF NOT EXISTS pipeline_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    category TEXT NOT NULL, -- 'cost_optimization', 'performance', 'security'
    config JSONB NOT NULL,
    estimated_savings_percent INTEGER DEFAULT 0,
    estimated_cpu_impact DECIMAL(5,2) DEFAULT 0,
    estimated_memory_impact_mb INTEGER DEFAULT 0,
    ui_preview JSONB DEFAULT '{}', -- Visual preview configuration
    tags TEXT[] DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_pipeline_templates_category ON pipeline_templates(category);
CREATE INDEX idx_pipeline_templates_active ON pipeline_templates(is_active);

-- Cost analytics aggregation table
CREATE TABLE IF NOT EXISTS cost_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMP NOT NULL,
    period TEXT NOT NULL, -- '1h', '1d', '7d', '30d'
    total_cost DECIMAL(12,2) NOT NULL DEFAULT 0,
    total_savings DECIMAL(12,2) NOT NULL DEFAULT 0,
    metrics_count BIGINT NOT NULL DEFAULT 0,
    cardinality_total BIGINT NOT NULL DEFAULT 0,
    cost_by_service JSONB DEFAULT '{}',
    cost_by_namespace JSONB DEFAULT '{}',
    cost_by_metric_type JSONB DEFAULT '{}',
    top_cost_drivers JSONB DEFAULT '[]',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_cost_analytics_timestamp ON cost_analytics(timestamp DESC);
CREATE INDEX idx_cost_analytics_period ON cost_analytics(period);

-- Real-time metric flow tracking
CREATE TABLE IF NOT EXISTS metric_flow_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    snapshot_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    total_metrics_per_sec BIGINT NOT NULL DEFAULT 0,
    total_cost_per_minute DECIMAL(10,2) NOT NULL DEFAULT 0,
    top_metrics JSONB DEFAULT '[]', -- Array of top N metrics by cost
    by_service JSONB DEFAULT '{}',
    by_namespace JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_metric_flow_time ON metric_flow_snapshots(snapshot_time DESC);

-- Experiment wizard configurations
CREATE TABLE IF NOT EXISTS experiment_wizard_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id TEXT REFERENCES experiments(id) ON DELETE CASCADE,
    wizard_config JSONB NOT NULL, -- Stores the wizard choices
    created_by TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- UI activity tracking for better UX
CREATE TABLE IF NOT EXISTS ui_activity_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT,
    session_id TEXT,
    action TEXT NOT NULL, -- 'view_dashboard', 'create_experiment', 'deploy_pipeline', etc.
    details JSONB DEFAULT '{}',
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_ui_activity_user ON ui_activity_log(user_id);
CREATE INDEX idx_ui_activity_action ON ui_activity_log(action);
CREATE INDEX idx_ui_activity_timestamp ON ui_activity_log(timestamp DESC);

-- Insert default pipeline templates
INSERT INTO pipeline_templates (name, description, category, config, estimated_savings_percent, estimated_cpu_impact, estimated_memory_impact_mb, ui_preview) VALUES
('top-k-20', 'Keep only top 20 metrics by value', 'cost_optimization', 
 '{"processors": [{"type": "top_k", "config": {"k": 20, "metric": "value"}}]}', 
 72, 0.5, 10, 
 '{"processor_blocks": [{"type": "top_k", "name": "Top 20 Filter", "config": {"k": 20}}]}'),
 
('priority-sli-slo', 'Prioritize SLI/SLO metrics only', 'cost_optimization',
 '{"processors": [{"type": "filter", "config": {"keep_patterns": ["sli.*", "slo.*", "error.*", "latency.*"]}}]}',
 65, 0.3, 5,
 '{"processor_blocks": [{"type": "filter", "name": "SLI/SLO Filter", "config": {"mode": "priority"}}]}'),
 
('adaptive-sampling', 'Adaptive sampling based on value changes', 'performance',
 '{"processors": [{"type": "adaptive_sample", "config": {"base_rate": 0.1, "spike_detection": true}}]}',
 45, 1.2, 25,
 '{"processor_blocks": [{"type": "sample", "name": "Adaptive Sampler", "config": {"mode": "adaptive"}}]}'),
 
('namespace-aggregate', 'Aggregate metrics by namespace', 'cost_optimization',
 '{"processors": [{"type": "aggregate", "config": {"group_by": ["namespace"], "operations": ["sum", "avg", "max"]}}]}',
 80, 0.8, 15,
 '{"processor_blocks": [{"type": "aggregate", "name": "Namespace Aggregator", "config": {"level": "namespace"}}]}');

-- Create update trigger for updated_at columns
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_agent_ui_state_updated_at BEFORE UPDATE ON agent_ui_state
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_pipeline_templates_updated_at BEFORE UPDATE ON pipeline_templates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();