-- Create pipeline deployments table for direct pipeline deployments
CREATE TABLE IF NOT EXISTS pipeline_deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deployment_name VARCHAR(255) NOT NULL,
    pipeline_name VARCHAR(255) NOT NULL,
    namespace VARCHAR(255) NOT NULL,
    target_nodes JSONB NOT NULL,
    parameters JSONB,
    resources JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    phase VARCHAR(50),
    instances JSONB, -- {desired: N, ready: N, updated: N}
    metrics JSONB,   -- Latest metrics snapshot
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    created_by VARCHAR(255),
    
    -- Ensure unique deployment names per namespace
    CONSTRAINT unique_deployment_name_namespace UNIQUE(deployment_name, namespace)
);

-- Create indexes for efficient queries
CREATE INDEX idx_deployments_namespace ON pipeline_deployments(namespace) WHERE deleted_at IS NULL;
CREATE INDEX idx_deployments_status ON pipeline_deployments(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_deployments_pipeline ON pipeline_deployments(pipeline_name) WHERE deleted_at IS NULL;
CREATE INDEX idx_deployments_created_at ON pipeline_deployments(created_at DESC);

-- Create deployment history table for audit trail
CREATE TABLE IF NOT EXISTS pipeline_deployment_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deployment_id UUID NOT NULL REFERENCES pipeline_deployments(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL, -- created, updated, started, stopped, deleted
    previous_state JSONB,
    new_state JSONB,
    message TEXT,
    performed_by VARCHAR(255),
    performed_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_deployment_history_deployment ON pipeline_deployment_history(deployment_id);
CREATE INDEX idx_deployment_history_performed_at ON pipeline_deployment_history(performed_at DESC);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_pipeline_deployments_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_pipeline_deployments_updated_at_trigger
    BEFORE UPDATE ON pipeline_deployments
    FOR EACH ROW
    EXECUTE FUNCTION update_pipeline_deployments_updated_at();

-- Create function to log deployment history
CREATE OR REPLACE FUNCTION log_pipeline_deployment_history()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO pipeline_deployment_history (
            deployment_id, action, new_state, performed_by
        ) VALUES (
            NEW.id, 'created', row_to_json(NEW), NEW.created_by
        );
    ELSIF TG_OP = 'UPDATE' THEN
        -- Only log if there's a meaningful change
        IF OLD.status != NEW.status OR OLD.phase != NEW.phase THEN
            INSERT INTO pipeline_deployment_history (
                deployment_id, action, previous_state, new_state, performed_by
            ) VALUES (
                NEW.id,
                CASE 
                    WHEN OLD.status != NEW.status THEN 'status_changed'
                    WHEN OLD.phase != NEW.phase THEN 'phase_changed'
                    ELSE 'updated'
                END,
                row_to_json(OLD),
                row_to_json(NEW),
                current_setting('app.current_user', true)
            );
        END IF;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO pipeline_deployment_history (
            deployment_id, action, previous_state, performed_by
        ) VALUES (
            OLD.id, 'deleted', row_to_json(OLD), current_setting('app.current_user', true)
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for deployment history
CREATE TRIGGER log_pipeline_deployment_history_trigger
    AFTER INSERT OR UPDATE OR DELETE ON pipeline_deployments
    FOR EACH ROW
    EXECUTE FUNCTION log_pipeline_deployment_history();

-- Sample data for testing (commented out)
-- INSERT INTO pipeline_deployments (
--     deployment_name,
--     pipeline_name,
--     namespace,
--     target_nodes,
--     parameters,
--     status,
--     phase
-- ) VALUES (
--     'production-topk',
--     'process-topk-v1',
--     'phoenix-prod',
--     '{"environment": "production", "tier": "frontend"}',
--     '{"top_k": 20, "critical_processes": ["nginx", "envoy"]}',
--     'active',
--     'running'
-- );