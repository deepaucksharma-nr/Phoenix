-- Authentication and Authorization Schema
-- Version: 1.0.0

SET search_path TO phoenix, public;

-- Create role enum
CREATE TYPE user_role AS ENUM ('admin', 'user', 'viewer');

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    tenant_id VARCHAR(255),
    roles user_role[] NOT NULL DEFAULT ARRAY['user']::user_role[],
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- API keys table for service-to-service auth
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL UNIQUE,
    service_name VARCHAR(255),
    permissions JSONB NOT NULL DEFAULT '[]',
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP WITH TIME ZONE
);

-- Session tokens table (for token blacklisting if needed)
CREATE TABLE revoked_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    user_id UUID REFERENCES users(id),
    revoked_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    reason TEXT
);

-- Tenants table
CREATE TABLE tenants (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    config JSONB DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add foreign key constraints
ALTER TABLE users ADD CONSTRAINT fk_users_tenant 
    FOREIGN KEY (tenant_id) REFERENCES tenants(id);

ALTER TABLE experiments ADD CONSTRAINT fk_experiments_created_by 
    FOREIGN KEY (created_by) REFERENCES users(id);

ALTER TABLE control_signals ADD CONSTRAINT fk_control_signals_applied_by 
    FOREIGN KEY (applied_by) REFERENCES users(id);

ALTER TABLE config_templates ADD CONSTRAINT fk_config_templates_created_by 
    FOREIGN KEY (created_by) REFERENCES users(id);

-- Create indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_active ON users(is_active) WHERE is_active = TRUE;

CREATE INDEX idx_api_keys_service_name ON api_keys(service_name);
CREATE INDEX idx_api_keys_expires_at ON api_keys(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX idx_api_keys_active ON api_keys(revoked_at) WHERE revoked_at IS NULL;

CREATE INDEX idx_revoked_tokens_expires_at ON revoked_tokens(expires_at);
CREATE INDEX idx_revoked_tokens_user_id ON revoked_tokens(user_id);

-- Update timestamp triggers
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tenants_updated_at BEFORE UPDATE ON tenants
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Audit triggers
CREATE TRIGGER audit_users AFTER INSERT OR UPDATE OR DELETE ON users
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_api_keys AFTER INSERT OR UPDATE OR DELETE ON api_keys
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_tenants AFTER INSERT OR UPDATE OR DELETE ON tenants
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

-- Helper functions
CREATE OR REPLACE FUNCTION check_user_permission(
    p_user_id UUID,
    p_resource VARCHAR,
    p_action VARCHAR
) RETURNS BOOLEAN AS $$
DECLARE
    v_roles user_role[];
BEGIN
    SELECT roles INTO v_roles FROM users WHERE id = p_user_id AND is_active = TRUE;
    
    -- Admin can do everything
    IF 'admin' = ANY(v_roles) THEN
        RETURN TRUE;
    END IF;
    
    -- Define permissions based on roles
    CASE p_resource
        WHEN 'experiment' THEN
            CASE p_action
                WHEN 'read' THEN
                    RETURN 'user' = ANY(v_roles) OR 'viewer' = ANY(v_roles);
                WHEN 'write' THEN
                    RETURN 'user' = ANY(v_roles);
                ELSE
                    RETURN FALSE;
            END CASE;
        WHEN 'template' THEN
            RETURN p_action = 'read' AND ('user' = ANY(v_roles) OR 'viewer' = ANY(v_roles));
        ELSE
            RETURN FALSE;
    END CASE;
END;
$$ LANGUAGE plpgsql;

-- Insert default tenant and admin user (password: admin123)
INSERT INTO tenants (id, name) VALUES 
    ('default', 'Default Tenant');

INSERT INTO users (email, password_hash, full_name, tenant_id, roles) VALUES 
    ('admin@phoenix.io', '$2a$10$YourHashedPasswordHere', 'Phoenix Admin', 'default', ARRAY['admin']::user_role[]);

-- Add comments
COMMENT ON TABLE users IS 'System users with authentication credentials';
COMMENT ON TABLE api_keys IS 'API keys for service-to-service authentication';
COMMENT ON TABLE revoked_tokens IS 'Blacklist for revoked JWT tokens';
COMMENT ON TABLE tenants IS 'Multi-tenancy support for the platform';
COMMENT ON FUNCTION check_user_permission IS 'Helper function to check user permissions';