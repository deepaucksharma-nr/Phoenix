-- Drop agent-related tables and modifications
DROP TABLE IF EXISTS agent_resources;
DROP TABLE IF EXISTS agent_capabilities;

-- Remove agent_id from tasks if it exists
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name='tasks' AND column_name='agent_id') THEN
        ALTER TABLE tasks DROP CONSTRAINT IF EXISTS fk_tasks_agent;
        ALTER TABLE tasks DROP COLUMN agent_id;
    END IF;
END $$;

DROP TABLE IF EXISTS agents;

-- Drop trigger and function
DROP TRIGGER IF EXISTS agents_updated_at_trigger ON agents;
DROP FUNCTION IF EXISTS update_agents_updated_at();