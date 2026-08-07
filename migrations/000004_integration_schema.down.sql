-- Rollback: 000004_integration_schema

DROP TABLE IF EXISTS analytics.channels;
DROP TABLE IF EXISTS integration.api_tokens;
DROP TABLE IF EXISTS integration.api_connections;
ALTER TABLE core.users DROP COLUMN IF EXISTS google_id;
DROP SCHEMA IF EXISTS integration CASCADE;
