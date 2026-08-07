-- Migration: 000008_api_tokens_unique_connection (rollback)
DROP INDEX IF EXISTS integration.uq_api_tokens_connection_id;
