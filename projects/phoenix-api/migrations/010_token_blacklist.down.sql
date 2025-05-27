-- Drop token blacklist table and related objects
DROP FUNCTION IF EXISTS cleanup_expired_tokens();
DROP TABLE IF EXISTS token_blacklist;