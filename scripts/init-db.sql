-- Phoenix Platform Database Initialization Script

-- Create databases if they don't exist
CREATE DATABASE IF NOT EXISTS phoenix_db;
CREATE DATABASE IF NOT EXISTS experiments_db;
CREATE DATABASE IF NOT EXISTS pipelines_db;

-- Create user if doesn't exist
DO
$$
BEGIN
   IF NOT EXISTS (
      SELECT FROM pg_catalog.pg_roles
      WHERE  rolname = 'phoenix') THEN
      CREATE ROLE phoenix LOGIN PASSWORD 'phoenix';
   END IF;
END
$$;

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE phoenix_db TO phoenix;
GRANT ALL PRIVILEGES ON DATABASE experiments_db TO phoenix;
GRANT ALL PRIVILEGES ON DATABASE pipelines_db TO phoenix;

-- Connect to phoenix_db and create extensions
\c phoenix_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Connect to experiments_db and create extensions
\c experiments_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Connect to pipelines_db and create extensions
\c pipelines_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";