-- Phoenix Platform Database Schema
-- Version: 1.0.0

-- Create schema
CREATE SCHEMA IF NOT EXISTS phoenix;

-- Set search path
SET search_path TO phoenix, public;

-- Create enum types
CREATE TYPE experiment_state AS ENUM (
    'pending',
    'initializing',
    'running',
    'analyzing',
    'completed',
    'failed',
    'cancelled'
);

CREATE TYPE signal_type AS ENUM (
    'traffic_split',
    'rollback',
    'config_update',
    'pause',
    'resume'
);

CREATE TYPE control_status AS ENUM (
    'pending',
    'active',
    'completed',
    'failed'
);

-- Experiments table
CREATE TABLE experiments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    state experiment_state NOT NULL DEFAULT 'pending',
    baseline_pipeline VARCHAR(255) NOT NULL,
    candidate_pipeline VARCHAR(255) NOT NULL,
    target_nodes JSONB NOT NULL DEFAULT '{}',
    config JSONB NOT NULL DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_by VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT experiments_name_unique UNIQUE (name, tenant_id)
);

-- Control signals table
CREATE TABLE control_signals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    type signal_type NOT NULL,
    parameters JSONB NOT NULL DEFAULT '{}',
    status control_status NOT NULL DEFAULT 'pending',
    reason TEXT,
    applied_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    applied_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT
);

-- Drift reports table
CREATE TABLE drift_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    drift_score DECIMAL(5,4) NOT NULL CHECK (drift_score >= 0 AND drift_score <= 1),
    metrics JSONB NOT NULL DEFAULT '[]',
    requires_action BOOLEAN NOT NULL DEFAULT FALSE,
    recommended_action TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Drift metrics table (for time series data)
CREATE TABLE drift_metrics (
    id BIGSERIAL PRIMARY KEY,
    experiment_id UUID NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    metric_name VARCHAR(255) NOT NULL,
    baseline_value DECIMAL NOT NULL,
    current_value DECIMAL NOT NULL,
    deviation_percentage DECIMAL NOT NULL,
    is_significant BOOLEAN NOT NULL DEFAULT FALSE,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Configuration templates table
CREATE TABLE config_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    content TEXT NOT NULL,
    version VARCHAR(50) NOT NULL DEFAULT 'v1.0.0',
    default_parameters JSONB DEFAULT '{}',
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Generated configurations table
CREATE TABLE generated_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    template_name VARCHAR(255) NOT NULL,
    configuration TEXT NOT NULL,
    parameters JSONB DEFAULT '{}',
    version VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Experiment results table
CREATE TABLE experiment_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    baseline_metrics JSONB NOT NULL,
    candidate_metrics JSONB NOT NULL,
    comparison JSONB NOT NULL,
    recommendation TEXT NOT NULL,
    raw_data JSONB DEFAULT '{}',
    analysis_timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Audit log table
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255),
    changes JSONB DEFAULT '{}',
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_experiments_state ON experiments(state);
CREATE INDEX idx_experiments_tenant_id ON experiments(tenant_id);
CREATE INDEX idx_experiments_created_at ON experiments(created_at DESC);

CREATE INDEX idx_control_signals_experiment_id ON control_signals(experiment_id);
CREATE INDEX idx_control_signals_status ON control_signals(status);
CREATE INDEX idx_control_signals_created_at ON control_signals(created_at DESC);

CREATE INDEX idx_drift_reports_experiment_id ON drift_reports(experiment_id);
CREATE INDEX idx_drift_reports_created_at ON drift_reports(created_at DESC);
CREATE INDEX idx_drift_reports_requires_action ON drift_reports(requires_action) WHERE requires_action = TRUE;

CREATE INDEX idx_drift_metrics_experiment_id ON drift_metrics(experiment_id);
CREATE INDEX idx_drift_metrics_timestamp ON drift_metrics(timestamp DESC);
CREATE INDEX idx_drift_metrics_significant ON drift_metrics(experiment_id, is_significant) WHERE is_significant = TRUE;

CREATE INDEX idx_generated_configs_experiment_id ON generated_configs(experiment_id);

CREATE INDEX idx_experiment_results_experiment_id ON experiment_results(experiment_id);

CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp DESC);

-- Create update timestamp trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_experiments_updated_at BEFORE UPDATE ON experiments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_config_templates_updated_at BEFORE UPDATE ON config_templates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create audit trigger
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO audit_logs(entity_type, entity_id, action, user_id, tenant_id, changes)
        VALUES (TG_TABLE_NAME, NEW.id, 'CREATE', current_setting('app.current_user', true), current_setting('app.current_tenant', true), row_to_json(NEW));
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO audit_logs(entity_type, entity_id, action, user_id, tenant_id, changes)
        VALUES (TG_TABLE_NAME, NEW.id, 'UPDATE', current_setting('app.current_user', true), current_setting('app.current_tenant', true), 
                jsonb_build_object('old', row_to_json(OLD), 'new', row_to_json(NEW)));
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO audit_logs(entity_type, entity_id, action, user_id, tenant_id, changes)
        VALUES (TG_TABLE_NAME, OLD.id, 'DELETE', current_setting('app.current_user', true), current_setting('app.current_tenant', true), row_to_json(OLD));
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ language 'plpgsql';

-- Apply audit triggers to main tables
CREATE TRIGGER audit_experiments AFTER INSERT OR UPDATE OR DELETE ON experiments
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_control_signals AFTER INSERT OR UPDATE OR DELETE ON control_signals
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_config_templates AFTER INSERT OR UPDATE OR DELETE ON config_templates
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

-- Add comments
COMMENT ON TABLE experiments IS 'Stores A/B testing experiments for OpenTelemetry pipeline optimization';
COMMENT ON TABLE control_signals IS 'Stores control signals applied to experiments';
COMMENT ON TABLE drift_reports IS 'Stores drift analysis reports for experiments';
COMMENT ON TABLE drift_metrics IS 'Time series data for drift metrics';
COMMENT ON TABLE config_templates IS 'Reusable configuration templates';
COMMENT ON TABLE generated_configs IS 'Generated configurations for experiments';
COMMENT ON TABLE experiment_results IS 'Final results and analysis for completed experiments';
COMMENT ON TABLE audit_logs IS 'Audit trail for all changes to the system';