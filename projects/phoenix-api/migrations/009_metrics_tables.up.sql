-- Create metrics table for storing time-series data
CREATE TABLE IF NOT EXISTS metrics (
    id BIGSERIAL PRIMARY KEY,
    experiment_id VARCHAR(255),
    source_id VARCHAR(255) NOT NULL,
    metric_name VARCHAR(255) NOT NULL,
    metric_type VARCHAR(50) NOT NULL, -- 'gauge', 'counter', 'histogram', 'cardinality'
    value DOUBLE PRECISION NOT NULL,
    labels JSONB DEFAULT '{}',
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for efficient queries
CREATE INDEX idx_metrics_timestamp ON metrics(timestamp DESC);
CREATE INDEX idx_metrics_experiment_id ON metrics(experiment_id) WHERE experiment_id IS NOT NULL;
CREATE INDEX idx_metrics_source_metric ON metrics(source_id, metric_name);
CREATE INDEX idx_metrics_labels_gin ON metrics USING gin(labels);
CREATE INDEX idx_metrics_type ON metrics(metric_type);

-- Create metric_cache table for caching recent metrics
CREATE TABLE IF NOT EXISTS metric_cache (
    id BIGSERIAL PRIMARY KEY,
    experiment_id VARCHAR(255),
    host_id VARCHAR(255) NOT NULL,
    metric_name VARCHAR(255) NOT NULL,
    variant VARCHAR(50),
    value DOUBLE PRECISION NOT NULL,
    labels JSONB DEFAULT '{}',
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_metric_cache_experiment ON metric_cache(experiment_id, timestamp DESC);
CREATE INDEX idx_metric_cache_host ON metric_cache(host_id, timestamp DESC);

-- Create cardinality_analysis table for storing cardinality data
CREATE TABLE IF NOT EXISTS cardinality_analysis (
    id BIGSERIAL PRIMARY KEY,
    metric_name VARCHAR(255) NOT NULL,
    label_name VARCHAR(255),
    unique_values INTEGER NOT NULL,
    total_series BIGINT NOT NULL,
    labels JSONB DEFAULT '{}',
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cardinality_timestamp ON cardinality_analysis(timestamp DESC);
CREATE INDEX idx_cardinality_metric ON cardinality_analysis(metric_name);
CREATE INDEX idx_cardinality_labels_gin ON cardinality_analysis USING gin(labels);

-- Create cost_tracking table for tracking costs over time
CREATE TABLE IF NOT EXISTS cost_tracking (
    id BIGSERIAL PRIMARY KEY,
    period VARCHAR(50) NOT NULL, -- 'hourly', 'daily', 'monthly'
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    total_metrics BIGINT NOT NULL,
    total_cost DECIMAL(10, 2) NOT NULL,
    cost_breakdown JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cost_tracking_period ON cost_tracking(period, start_time DESC);

-- Create pipeline_templates table to replace hardcoded templates
CREATE TABLE IF NOT EXISTS pipeline_templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    version VARCHAR(50) NOT NULL,
    config_url TEXT NOT NULL,
    tags TEXT[] DEFAULT '{}',
    estimated_reduction DECIMAL(5, 2) DEFAULT 0.0,
    features TEXT[] DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_pipeline_templates_name ON pipeline_templates(name);
CREATE INDEX idx_pipeline_templates_tags ON pipeline_templates USING gin(tags);

-- Insert default pipeline templates
INSERT INTO pipeline_templates (name, display_name, description, version, config_url, tags, estimated_reduction, features, metadata) VALUES
('adaptive-filter-v2', 'Adaptive Filter v2.0', 'Intelligent metric filtering with ML-based cardinality reduction', '2.0.3', 's3://phoenix-configs/optimized/adaptive-filter.yaml', 
 ARRAY['production', 'ml', 'high-efficiency'], 70.0, 
 ARRAY['Dynamic sampling', 'Pattern recognition', 'Auto-tuning'],
 '{"category": "optimization", "maturity": "stable", "risk": "low"}'::jsonb),
 
('topk-aggregator', 'Top-K Aggregator', 'Keep only the most important metrics using frequency analysis', '1.5.1', 's3://phoenix-configs/optimized/topk-config.yaml',
 ARRAY['production', 'simple', 'reliable'], 60.0,
 ARRAY['Frequency analysis', 'Threshold tuning', 'Stable output'],
 '{"category": "aggregation", "maturity": "stable", "risk": "low"}'::jsonb),
 
('current-production', 'Current Production', 'Baseline production configuration with minimal filtering', '1.0.0', 's3://phoenix-configs/baseline/production.yaml',
 ARRAY['baseline', 'production', 'no-optimization'], 0.0,
 ARRAY['Full metrics', 'No filtering', 'Maximum visibility'],
 '{"category": "baseline", "maturity": "stable", "risk": "none"}'::jsonb);

-- Create function to calculate metric costs
CREATE OR REPLACE FUNCTION calculate_metric_cost(
    cardinality BIGINT,
    rate_per_million DECIMAL DEFAULT 50.0
) RETURNS DECIMAL AS $$
BEGIN
    -- Calculate cost based on cardinality
    -- Default: $50 per million metrics per month = $0.00167 per million per minute
    RETURN (cardinality::DECIMAL / 1000000.0) * (rate_per_million / 30.0 / 24.0 / 60.0);
END;
$$ LANGUAGE plpgsql;

-- Create view for real-time cost flow
CREATE OR REPLACE VIEW metric_cost_flow_view AS
WITH recent_metrics AS (
    SELECT 
        metric_name,
        labels,
        COUNT(*) as data_points,
        COUNT(DISTINCT labels) as cardinality
    FROM metrics
    WHERE timestamp > NOW() - INTERVAL '5 minutes'
    GROUP BY metric_name, labels
),
cost_calculation AS (
    SELECT 
        metric_name,
        labels,
        cardinality,
        calculate_metric_cost(cardinality, 50.0) as cost_per_minute
    FROM recent_metrics
)
SELECT 
    metric_name,
    labels,
    cardinality,
    cost_per_minute,
    (cost_per_minute / SUM(cost_per_minute) OVER () * 100) as percentage,
    SUM(cost_per_minute) OVER () as total_cost
FROM cost_calculation
ORDER BY cost_per_minute DESC;