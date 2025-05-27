-- Create metrics table for caching agent metrics
CREATE TABLE IF NOT EXISTS metrics (
    id BIGSERIAL PRIMARY KEY,
    source_id VARCHAR(255) NOT NULL, -- host_id or agent_id
    metric_name VARCHAR(255) NOT NULL,
    metric_type VARCHAR(50) NOT NULL DEFAULT 'gauge',
    value DOUBLE PRECISION NOT NULL,
    labels JSONB NOT NULL DEFAULT '{}',
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Unique constraint to prevent duplicates
    CONSTRAINT unique_metric_point UNIQUE (source_id, metric_name, labels, timestamp)
);

-- Create indexes for faster queries
CREATE INDEX idx_metrics_source_timestamp ON metrics(source_id, timestamp DESC);
CREATE INDEX idx_metrics_name_timestamp ON metrics(metric_name, timestamp DESC);
CREATE INDEX idx_metrics_labels ON metrics USING GIN (labels);
CREATE INDEX idx_metrics_timestamp ON metrics(timestamp DESC);

-- Create cardinality analysis table for tracking unique label values
CREATE TABLE IF NOT EXISTS cardinality_analysis (
    id BIGSERIAL PRIMARY KEY,
    metric_name VARCHAR(255) NOT NULL,
    label_name VARCHAR(255) NOT NULL,
    label_value TEXT NOT NULL,
    source_id VARCHAR(255) NOT NULL,
    unique_values INTEGER DEFAULT 1,
    total_series INTEGER DEFAULT 1,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Unique constraint
    CONSTRAINT unique_cardinality_entry UNIQUE (metric_name, label_name, label_value, source_id)
);

-- Create indexes for cardinality analysis
CREATE INDEX idx_cardinality_metric_label ON cardinality_analysis(metric_name, label_name);
CREATE INDEX idx_cardinality_timestamp ON cardinality_analysis(timestamp DESC);
CREATE INDEX idx_cardinality_source ON cardinality_analysis(source_id);

-- Create a materialized view for fast cardinality queries
CREATE MATERIALIZED VIEW IF NOT EXISTS cardinality_summary AS
SELECT 
    metric_name,
    label_name,
    COUNT(DISTINCT label_value) as unique_values,
    COUNT(*) as total_series,
    MAX(timestamp) as last_updated
FROM cardinality_analysis
GROUP BY metric_name, label_name;

-- Create index on materialized view
CREATE INDEX idx_cardinality_summary_metric ON cardinality_summary(metric_name);

-- Create function to refresh materialized view (can be called periodically)
CREATE OR REPLACE FUNCTION refresh_cardinality_summary()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY cardinality_summary;
END;
$$ LANGUAGE plpgsql;