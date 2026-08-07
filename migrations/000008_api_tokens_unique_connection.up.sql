-- Migration: 000008_api_tokens_unique_connection
-- Description: Fix missing unique constraint on api_tokens.connection_id (required by ON CONFLICT in SaveToken)

CREATE UNIQUE INDEX IF NOT EXISTS uq_api_tokens_connection_id
    ON integration.api_tokens(connection_id);
