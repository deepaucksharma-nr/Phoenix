-- Create deployment versions table to track deployment history
CREATE TABLE IF NOT EXISTS deployment_versions (
    id BIGSERIAL PRIMARY KEY,
    deployment_id VARCHAR(255) NOT NULL,
    version INTEGER NOT NULL,
    pipeline_config TEXT NOT NULL, -- Stored rendered pipeline configuration
    parameters JSONB NOT NULL DEFAULT '{}',
    deployed_by VARCHAR(255),
    deployed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, superseded, rolled_back
    rollback_from_version INTEGER, -- If this is a rollback, which version it rolled back from
    notes TEXT,
    
    -- Unique constraint on deployment_id and version
    CONSTRAINT unique_deployment_version UNIQUE (deployment_id, version)
);

-- Create indexes for faster queries
CREATE INDEX idx_deployment_versions_deployment ON deployment_versions(deployment_id, version DESC);
CREATE INDEX idx_deployment_versions_status ON deployment_versions(status);
CREATE INDEX idx_deployment_versions_deployed_at ON deployment_versions(deployed_at DESC);

-- Add version column to pipeline_deployments table
ALTER TABLE pipeline_deployments 
ADD COLUMN IF NOT EXISTS current_version INTEGER DEFAULT 1,
ADD COLUMN IF NOT EXISTS last_version_change TIMESTAMP;

-- Create function to get next version number
CREATE OR REPLACE FUNCTION get_next_deployment_version(p_deployment_id VARCHAR)
RETURNS INTEGER AS $$
DECLARE
    next_version INTEGER;
BEGIN
    SELECT COALESCE(MAX(version), 0) + 1 INTO next_version
    FROM deployment_versions
    WHERE deployment_id = p_deployment_id;
    
    RETURN next_version;
END;
$$ LANGUAGE plpgsql;

-- Create function to record deployment version
CREATE OR REPLACE FUNCTION record_deployment_version(
    p_deployment_id VARCHAR,
    p_pipeline_config TEXT,
    p_parameters JSONB,
    p_deployed_by VARCHAR,
    p_notes TEXT DEFAULT NULL
)
RETURNS INTEGER AS $$
DECLARE
    new_version INTEGER;
BEGIN
    -- Get next version number
    new_version := get_next_deployment_version(p_deployment_id);
    
    -- Mark previous versions as superseded
    UPDATE deployment_versions 
    SET status = 'superseded' 
    WHERE deployment_id = p_deployment_id 
    AND status = 'active';
    
    -- Insert new version
    INSERT INTO deployment_versions (
        deployment_id, version, pipeline_config, parameters, 
        deployed_by, status, notes
    ) VALUES (
        p_deployment_id, new_version, p_pipeline_config, p_parameters,
        p_deployed_by, 'active', p_notes
    );
    
    -- Update deployment table
    UPDATE pipeline_deployments 
    SET current_version = new_version,
        last_version_change = CURRENT_TIMESTAMP
    WHERE id = p_deployment_id;
    
    RETURN new_version;
END;
$$ LANGUAGE plpgsql;

-- Create function for rollback
CREATE OR REPLACE FUNCTION rollback_deployment_version(
    p_deployment_id VARCHAR,
    p_target_version INTEGER,
    p_rolled_back_by VARCHAR,
    p_notes TEXT DEFAULT NULL
)
RETURNS BOOLEAN AS $$
DECLARE
    current_version INTEGER;
    rollback_config TEXT;
    rollback_params JSONB;
BEGIN
    -- Get current version
    SELECT current_version INTO current_version
    FROM pipeline_deployments
    WHERE id = p_deployment_id;
    
    -- Get target version config
    SELECT pipeline_config, parameters 
    INTO rollback_config, rollback_params
    FROM deployment_versions
    WHERE deployment_id = p_deployment_id 
    AND version = p_target_version;
    
    IF rollback_config IS NULL THEN
        RETURN FALSE;
    END IF;
    
    -- Create new version as rollback
    INSERT INTO deployment_versions (
        deployment_id, version, pipeline_config, parameters,
        deployed_by, status, rollback_from_version, notes
    ) VALUES (
        p_deployment_id, 
        get_next_deployment_version(p_deployment_id),
        rollback_config, 
        rollback_params,
        p_rolled_back_by, 
        'active', 
        current_version,
        COALESCE(p_notes, 'Rollback to version ' || p_target_version)
    );
    
    -- Mark current version as rolled back
    UPDATE deployment_versions 
    SET status = 'rolled_back' 
    WHERE deployment_id = p_deployment_id 
    AND version = current_version;
    
    -- Update deployment table
    UPDATE pipeline_deployments 
    SET current_version = get_next_deployment_version(p_deployment_id) - 1,
        last_version_change = CURRENT_TIMESTAMP
    WHERE id = p_deployment_id;
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;