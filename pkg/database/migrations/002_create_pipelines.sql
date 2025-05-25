-- 002_create_pipelines.sql
-- Pipeline configurations and templates

-- Create enum for pipeline status
CREATE TYPE pipeline_status AS ENUM (
    'draft',
    'validated',
    'deployed',
    'active',
    'deprecated',
    'failed'
);

-- Create enum for pipeline type
CREATE TYPE pipeline_type AS ENUM (
    'baseline',
    'optimized',
    'custom'
);

-- Pipeline templates catalog
CREATE TABLE IF NOT EXISTS pipeline_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    version VARCHAR(50) NOT NULL,
    type pipeline_type NOT NULL,
    
    -- Configuration
    config_yaml TEXT NOT NULL,
    config_json JSONB NOT NULL,
    
    -- Optimization metrics
    expected_reduction_percent INTEGER CHECK (expected_reduction_percent >= 0 AND expected_reduction_percent <= 100),
    complexity_score INTEGER CHECK (complexity_score >= 1 AND complexity_score <= 10),
    
    -- Metadata
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Pipeline instances (deployed configurations)
CREATE TABLE IF NOT EXISTS pipelines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    template_id UUID REFERENCES pipeline_templates(id),
    experiment_id UUID REFERENCES experiments(id) ON DELETE CASCADE,
    
    -- Configuration
    config_yaml TEXT NOT NULL,
    config_json JSONB NOT NULL,
    config_hash VARCHAR(64) NOT NULL, -- SHA256 of config
    
    -- Deployment info
    status pipeline_status NOT NULL DEFAULT 'draft',
    k8s_namespace VARCHAR(255),
    k8s_resource_name VARCHAR(255),
    
    -- Variables used in template
    variables JSONB NOT NULL DEFAULT '{}',
    
    -- Metrics
    deployment_count INTEGER DEFAULT 0,
    last_deployed_at TIMESTAMP WITH TIME ZONE,
    
    -- Metadata
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Soft delete
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Pipeline deployment history
CREATE TABLE IF NOT EXISTS pipeline_deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pipeline_id UUID NOT NULL REFERENCES pipelines(id) ON DELETE CASCADE,
    experiment_id UUID REFERENCES experiments(id) ON DELETE CASCADE,
    
    -- Deployment details
    deployed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deployed_by VARCHAR(255) NOT NULL,
    deployment_type VARCHAR(50) NOT NULL, -- 'create', 'update', 'delete'
    
    -- Target information
    target_nodes JSONB NOT NULL DEFAULT '[]',
    target_count INTEGER DEFAULT 0,
    
    -- Status
    status VARCHAR(50) NOT NULL, -- 'pending', 'success', 'failed'
    error_message TEXT,
    
    -- Git information (for GitOps)
    git_commit_sha VARCHAR(40),
    git_pr_url TEXT,
    
    -- Metadata
    metadata JSONB DEFAULT '{}'
);

-- Indexes
CREATE INDEX idx_pipeline_templates_active ON pipeline_templates(is_active);
CREATE INDEX idx_pipelines_experiment ON pipelines(experiment_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_pipelines_status ON pipelines(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_pipelines_hash ON pipelines(config_hash);
CREATE INDEX idx_deployments_pipeline ON pipeline_deployments(pipeline_id);
CREATE INDEX idx_deployments_time ON pipeline_deployments(deployed_at DESC);

-- Triggers
CREATE TRIGGER update_pipeline_templates_updated_at BEFORE UPDATE
    ON pipeline_templates FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_pipelines_updated_at BEFORE UPDATE
    ON pipelines FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default pipeline templates
INSERT INTO pipeline_templates (name, description, version, type, config_yaml, config_json, expected_reduction_percent, complexity_score) VALUES
('process-baseline-v1', 'Baseline configuration with no optimization', 'v1.0.0', 'baseline', 
'receivers:
  hostmetrics:
    collection_interval: 10s
    scrapers:
      process:
        include:
          match_type: regexp
          names: [".*"]
processors:
  batch:
    timeout: 10s
exporters:
  otlp:
    endpoint: "${OTLP_ENDPOINT}"
service:
  pipelines:
    metrics:
      receivers: [hostmetrics]
      processors: [batch]
      exporters: [otlp]',
'{"receivers":{"hostmetrics":{"collection_interval":"10s","scrapers":{"process":{"include":{"match_type":"regexp","names":[".*"]}}}}},"processors":{"batch":{"timeout":"10s"}},"exporters":{"otlp":{"endpoint":"${OTLP_ENDPOINT}"}},"service":{"pipelines":{"metrics":{"receivers":["hostmetrics"],"processors":["batch"],"exporters":["otlp"]}}}}',
0, 1),

('process-priority-filter-v1', 'Filter processes by priority/importance', 'v1.0.0', 'optimized',
'receivers:
  hostmetrics:
    collection_interval: 30s
    scrapers:
      process:
        include:
          match_type: regexp
          names: ["critical-.*", "important-.*"]
processors:
  batch:
    timeout: 30s
  filter:
    metrics:
      include:
        match_type: expr
        expressions:
          - ''process.cpu.utilization > 0.05''
exporters:
  otlp:
    endpoint: "${OTLP_ENDPOINT}"
service:
  pipelines:
    metrics:
      receivers: [hostmetrics]
      processors: [filter, batch]
      exporters: [otlp]',
'{"receivers":{"hostmetrics":{"collection_interval":"30s","scrapers":{"process":{"include":{"match_type":"regexp","names":["critical-.*","important-.*"]}}}}},"processors":{"batch":{"timeout":"30s"},"filter":{"metrics":{"include":{"match_type":"expr","expressions":["process.cpu.utilization > 0.05"]}}}},"exporters":{"otlp":{"endpoint":"${OTLP_ENDPOINT}"}},"service":{"pipelines":{"metrics":{"receivers":["hostmetrics"],"processors":["filter","batch"],"exporters":["otlp"]}}}}',
50, 3);

-- Comments
COMMENT ON TABLE pipeline_templates IS 'Catalog of pre-validated pipeline configurations';
COMMENT ON TABLE pipelines IS 'Pipeline instances created from templates';
COMMENT ON TABLE pipeline_deployments IS 'History of pipeline deployments to clusters';