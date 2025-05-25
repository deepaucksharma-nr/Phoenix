-- 003_create_experiment_states.sql
-- Track experiment state transitions and events

-- State transition history
CREATE TABLE IF NOT EXISTS experiment_state_transitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    
    -- State change
    from_state experiment_status,
    to_state experiment_status NOT NULL,
    
    -- Transition details
    triggered_by VARCHAR(255) NOT NULL, -- 'system', 'user', 'scheduler', 'timeout'
    triggered_by_user VARCHAR(255),
    reason TEXT,
    
    -- Timing
    transitioned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Additional context
    metadata JSONB DEFAULT '{}'
);

-- Experiment events log
CREATE TABLE IF NOT EXISTS experiment_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    
    -- Event information
    event_type VARCHAR(100) NOT NULL,
    event_source VARCHAR(100) NOT NULL, -- 'controller', 'operator', 'api', 'user'
    severity VARCHAR(20) NOT NULL DEFAULT 'info', -- 'info', 'warning', 'error', 'critical'
    
    -- Event details
    message TEXT NOT NULL,
    details JSONB DEFAULT '{}',
    
    -- Timing
    occurred_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- User context
    user_id VARCHAR(255)
);

-- Experiment metrics snapshots
CREATE TABLE IF NOT EXISTS experiment_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    
    -- Metric identification
    pipeline_type VARCHAR(20) NOT NULL, -- 'baseline' or 'candidate'
    metric_name VARCHAR(255) NOT NULL,
    
    -- Metric values
    value DOUBLE PRECISION NOT NULL,
    unit VARCHAR(50),
    
    -- Time window
    window_start TIMESTAMP WITH TIME ZONE NOT NULL,
    window_end TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Aggregation info
    aggregation_type VARCHAR(20) NOT NULL DEFAULT 'avg', -- 'avg', 'sum', 'min', 'max', 'p50', 'p95', 'p99'
    sample_count INTEGER,
    
    -- Collection time
    collected_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Additional labels/dimensions
    labels JSONB DEFAULT '{}'
);

-- Experiment analysis results
CREATE TABLE IF NOT EXISTS experiment_analysis (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    
    -- Analysis metadata
    analysis_type VARCHAR(50) NOT NULL, -- 'statistical', 'ml_based', 'rule_based'
    analyzer_version VARCHAR(20) NOT NULL,
    
    -- Results
    baseline_score DOUBLE PRECISION,
    candidate_score DOUBLE PRECISION,
    confidence_level DOUBLE PRECISION, -- 0-1
    p_value DOUBLE PRECISION,
    
    -- Recommendations
    recommendation VARCHAR(20) NOT NULL, -- 'promote_baseline', 'promote_candidate', 'continue', 'stop'
    reasoning TEXT,
    
    -- Detailed metrics comparison
    metrics_comparison JSONB NOT NULL DEFAULT '{}',
    
    -- Timing
    analyzed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    analysis_duration_ms INTEGER
);

-- Indexes
CREATE INDEX idx_state_transitions_experiment ON experiment_state_transitions(experiment_id, transitioned_at DESC);
CREATE INDEX idx_state_transitions_states ON experiment_state_transitions(from_state, to_state);
CREATE INDEX idx_events_experiment ON experiment_events(experiment_id, occurred_at DESC);
CREATE INDEX idx_events_type ON experiment_events(event_type, severity);
CREATE INDEX idx_metrics_experiment ON experiment_metrics(experiment_id, metric_name, collected_at DESC);
CREATE INDEX idx_metrics_window ON experiment_metrics(window_start, window_end);
CREATE INDEX idx_analysis_experiment ON experiment_analysis(experiment_id, analyzed_at DESC);

-- Create function to validate state transitions
CREATE OR REPLACE FUNCTION validate_state_transition()
RETURNS TRIGGER AS $$
DECLARE
    valid_transition BOOLEAN := FALSE;
BEGIN
    -- Define valid state transitions
    CASE NEW.from_state
        WHEN 'draft' THEN
            valid_transition := NEW.to_state IN ('pending', 'cancelled');
        WHEN 'pending' THEN
            valid_transition := NEW.to_state IN ('initializing', 'cancelled');
        WHEN 'initializing' THEN
            valid_transition := NEW.to_state IN ('deploying', 'failed', 'cancelled');
        WHEN 'deploying' THEN
            valid_transition := NEW.to_state IN ('running', 'failed', 'cancelled');
        WHEN 'running' THEN
            valid_transition := NEW.to_state IN ('analyzing', 'completed', 'failed', 'cancelled');
        WHEN 'analyzing' THEN
            valid_transition := NEW.to_state IN ('completed', 'failed');
        WHEN 'completed', 'failed', 'cancelled' THEN
            valid_transition := FALSE; -- Terminal states
        ELSE
            valid_transition := TRUE; -- Allow if from_state is NULL (initial state)
    END CASE;
    
    IF NOT valid_transition THEN
        RAISE EXCEPTION 'Invalid state transition from % to %', NEW.from_state, NEW.to_state;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';

CREATE TRIGGER validate_experiment_state_transition
    BEFORE INSERT ON experiment_state_transitions
    FOR EACH ROW EXECUTE FUNCTION validate_state_transition();

-- Comments
COMMENT ON TABLE experiment_state_transitions IS 'Audit log of all experiment state changes';
COMMENT ON TABLE experiment_events IS 'Event log for experiment lifecycle events';
COMMENT ON TABLE experiment_metrics IS 'Time-series metrics collected during experiments';
COMMENT ON TABLE experiment_analysis IS 'Analysis results and recommendations for experiments';