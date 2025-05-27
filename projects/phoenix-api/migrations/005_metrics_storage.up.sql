-- Metrics storage for experiments and pipelines
CREATE TABLE IF NOT EXISTS metrics (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    source_type VARCHAR(50) NOT NULL, -- 'experiment', 'pipeline', 'agent'
    source_id VARCHAR(255) NOT NULL,
    metric_type VARCHAR(100) NOT NULL, -- 'cardinality', 'cpu', 'memory', 'throughput', etc
    metric_name VARCHAR(255) NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    unit VARCHAR(50), -- 'percent', 'bytes', 'count', 'ms', etc
    labels JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}'
);

-- Indexes for efficient querying
CREATE INDEX idx_metrics_timestamp ON metrics(timestamp DESC);
CREATE INDEX idx_metrics_source ON metrics(source_type, source_id);
CREATE INDEX idx_metrics_type_name ON metrics(metric_type, metric_name);
CREATE INDEX idx_metrics_source_timestamp ON metrics(source_type, source_id, timestamp DESC);

-- Create hypertable for time-series data if TimescaleDB is available
-- COMMENTED OUT: Uncomment if using TimescaleDB
-- SELECT create_hypertable('metrics', 'timestamp', if_not_exists => TRUE);

-- Aggregated metrics for faster queries
CREATE TABLE IF NOT EXISTS metrics_aggregated (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid(),
    source_type VARCHAR(50) NOT NULL,
    source_id VARCHAR(255) NOT NULL,
    metric_type VARCHAR(100) NOT NULL,
    metric_name VARCHAR(255) NOT NULL,
    period VARCHAR(20) NOT NULL, -- '1m', '5m', '1h', '1d'
    timestamp TIMESTAMP NOT NULL,
    count INTEGER NOT NULL,
    sum DOUBLE PRECISION NOT NULL,
    min DOUBLE PRECISION NOT NULL,
    max DOUBLE PRECISION NOT NULL,
    avg DOUBLE PRECISION NOT NULL,
    p50 DOUBLE PRECISION,
    p90 DOUBLE PRECISION,
    p95 DOUBLE PRECISION,
    p99 DOUBLE PRECISION,
    stddev DOUBLE PRECISION,
    metadata JSONB DEFAULT '{}'
);

-- Indexes for aggregated metrics
CREATE INDEX idx_metrics_agg_source_period ON metrics_aggregated(source_type, source_id, period, timestamp DESC);
CREATE UNIQUE INDEX idx_metrics_agg_unique ON metrics_aggregated(source_type, source_id, metric_type, metric_name, period, timestamp);

-- Cost tracking table
CREATE TABLE IF NOT EXISTS cost_tracking (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    source_type VARCHAR(50) NOT NULL,
    source_id VARCHAR(255) NOT NULL,
    cost_type VARCHAR(100) NOT NULL, -- 'compute', 'storage', 'network', 'total'
    amount DECIMAL(10, 4) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL,
    breakdown JSONB DEFAULT '{}', -- Detailed cost breakdown
    metadata JSONB DEFAULT '{}'
);

-- Indexes for cost tracking
CREATE INDEX idx_cost_tracking_source ON cost_tracking(source_type, source_id);
CREATE INDEX idx_cost_tracking_period ON cost_tracking(period_start, period_end);

-- Cardinality tracking for detailed analysis
CREATE TABLE IF NOT EXISTS cardinality_analysis (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    pipeline_id VARCHAR(255) NOT NULL,
    metric_name VARCHAR(255) NOT NULL,
    label_name VARCHAR(255) NOT NULL,
    unique_values INTEGER NOT NULL,
    top_values JSONB DEFAULT '[]', -- Array of {value, count} objects
    total_series INTEGER NOT NULL,
    reduction_potential DOUBLE PRECISION, -- Percentage that could be reduced
    metadata JSONB DEFAULT '{}'
);

-- Indexes for cardinality analysis
CREATE INDEX idx_cardinality_pipeline_timestamp ON cardinality_analysis(pipeline_id, timestamp DESC);
CREATE INDEX idx_cardinality_metric_label ON cardinality_analysis(metric_name, label_name);

-- Real-time metrics buffer for streaming
CREATE TABLE IF NOT EXISTS metrics_buffer (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid(),
    received_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    agent_id VARCHAR(255),
    batch_id VARCHAR(255),
    metrics JSONB NOT NULL, -- Array of metric objects
    processed BOOLEAN NOT NULL DEFAULT false,
    processed_at TIMESTAMP
);

-- Index for processing
CREATE INDEX idx_metrics_buffer_unprocessed ON metrics_buffer(processed, received_at) WHERE processed = false;

-- Function to clean old metrics
CREATE OR REPLACE FUNCTION clean_old_metrics(retention_days INTEGER DEFAULT 30)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM metrics
    WHERE timestamp < CURRENT_TIMESTAMP - INTERVAL '1 day' * retention_days;
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    DELETE FROM metrics_buffer
    WHERE processed = true AND processed_at < CURRENT_TIMESTAMP - INTERVAL '1 day';
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;