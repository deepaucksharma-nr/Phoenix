-- 004_create_metrics.sql
-- Platform metrics and monitoring data

-- Platform usage metrics
CREATE TABLE IF NOT EXISTS platform_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Metric identification
    metric_type VARCHAR(100) NOT NULL, -- 'api_request', 'experiment_duration', 'pipeline_deployment', etc.
    metric_name VARCHAR(255) NOT NULL,
    
    -- Metric value
    value DOUBLE PRECISION NOT NULL,
    unit VARCHAR(50),
    
    -- Context
    service_name VARCHAR(100) NOT NULL,
    operation_name VARCHAR(255),
    
    -- Dimensions
    labels JSONB DEFAULT '{}',
    
    -- Timing
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Request tracing
    trace_id VARCHAR(32),
    span_id VARCHAR(16)
);

-- User activity tracking
CREATE TABLE IF NOT EXISTS user_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- User identification
    user_id VARCHAR(255) NOT NULL,
    user_email VARCHAR(255),
    
    -- Activity details
    activity_type VARCHAR(100) NOT NULL, -- 'experiment_created', 'pipeline_deployed', 'analysis_viewed', etc.
    resource_type VARCHAR(50), -- 'experiment', 'pipeline', 'report'
    resource_id UUID,
    
    -- Activity context
    action VARCHAR(50) NOT NULL, -- 'create', 'read', 'update', 'delete'
    description TEXT,
    
    -- Request details
    ip_address INET,
    user_agent TEXT,
    
    -- Timing
    occurred_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Additional context
    metadata JSONB DEFAULT '{}'
);

-- Cost tracking
CREATE TABLE IF NOT EXISTS cost_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Cost identification
    experiment_id UUID REFERENCES experiments(id) ON DELETE SET NULL,
    pipeline_id UUID REFERENCES pipelines(id) ON DELETE SET NULL,
    
    -- Cost details
    metric_type VARCHAR(50) NOT NULL, -- 'data_points', 'api_calls', 'storage', 'compute'
    
    -- Before/after comparison
    baseline_value DOUBLE PRECISION,
    optimized_value DOUBLE PRECISION,
    reduction_percent DOUBLE PRECISION,
    
    -- Cost calculation
    unit_cost DECIMAL(10, 4),
    currency VARCHAR(3) DEFAULT 'USD',
    total_saved DECIMAL(10, 2),
    
    -- Time period
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Calculation metadata
    calculated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    calculation_method VARCHAR(50) -- 'estimated', 'actual', 'projected'
);

-- System health metrics
CREATE TABLE IF NOT EXISTS system_health (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Component identification
    component_name VARCHAR(100) NOT NULL, -- 'api', 'controller', 'operator', etc.
    instance_id VARCHAR(255) NOT NULL,
    
    -- Health status
    status VARCHAR(20) NOT NULL, -- 'healthy', 'degraded', 'unhealthy'
    
    -- Resource usage
    cpu_percent DOUBLE PRECISION,
    memory_bytes BIGINT,
    disk_bytes BIGINT,
    
    -- Performance metrics
    response_time_ms DOUBLE PRECISION,
    error_rate DOUBLE PRECISION,
    throughput DOUBLE PRECISION,
    
    -- Connectivity
    upstream_healthy BOOLEAN DEFAULT true,
    downstream_healthy BOOLEAN DEFAULT true,
    
    -- Timing
    checked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Additional checks
    health_checks JSONB DEFAULT '{}'
);

-- Indexes for performance
CREATE INDEX idx_platform_metrics_type_time ON platform_metrics(metric_type, timestamp DESC);
CREATE INDEX idx_platform_metrics_service ON platform_metrics(service_name, timestamp DESC);
CREATE INDEX idx_user_activities_user_time ON user_activities(user_id, occurred_at DESC);
CREATE INDEX idx_user_activities_type ON user_activities(activity_type, occurred_at DESC);
CREATE INDEX idx_cost_metrics_experiment ON cost_metrics(experiment_id, period_end DESC);
CREATE INDEX idx_cost_metrics_time ON cost_metrics(period_end DESC);
CREATE INDEX idx_system_health_component ON system_health(component_name, checked_at DESC);
CREATE INDEX idx_system_health_status ON system_health(status, checked_at DESC);

-- Partitioning for high-volume tables
-- Platform metrics partitioned by month
CREATE TABLE platform_metrics_2024_01 PARTITION OF platform_metrics
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
    
CREATE TABLE platform_metrics_2024_02 PARTITION OF platform_metrics
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');

-- Create function to auto-create monthly partitions
CREATE OR REPLACE FUNCTION create_monthly_partition()
RETURNS void AS $$
DECLARE
    start_date date;
    end_date date;
    partition_name text;
BEGIN
    start_date := date_trunc('month', CURRENT_DATE);
    end_date := start_date + interval '1 month';
    partition_name := 'platform_metrics_' || to_char(start_date, 'YYYY_MM');
    
    -- Check if partition exists
    IF NOT EXISTS (
        SELECT 1 FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE c.relname = partition_name
        AND n.nspname = 'public'
    ) THEN
        EXECUTE format('CREATE TABLE %I PARTITION OF platform_metrics FOR VALUES FROM (%L) TO (%L)',
            partition_name, start_date, end_date);
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Create a scheduled job to create partitions (requires pg_cron or similar)
-- SELECT cron.schedule('create-partitions', '0 0 1 * *', 'SELECT create_monthly_partition()');

-- Comments
COMMENT ON TABLE platform_metrics IS 'General platform metrics for monitoring and observability';
COMMENT ON TABLE user_activities IS 'User activity audit log';
COMMENT ON TABLE cost_metrics IS 'Cost tracking and savings calculations';
COMMENT ON TABLE system_health IS 'Component health and performance metrics';