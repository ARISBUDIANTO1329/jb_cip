-- Migration: 000004_integration_schema
-- Description: Add integration tables for Google OAuth and YouTube channels

-- Create schemas
CREATE SCHEMA IF NOT EXISTS integration;
CREATE SCHEMA IF NOT EXISTS analytics;

-- Add google_id to users table
ALTER TABLE core.users ADD COLUMN IF NOT EXISTS google_id TEXT UNIQUE;

-- Create API connections table
CREATE TABLE IF NOT EXISTS integration.api_connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES core.users(id) ON DELETE CASCADE,
    workspace_id UUID NOT NULL REFERENCES core.workspaces(id) ON DELETE CASCADE,
    provider TEXT NOT NULL DEFAULT 'google',
    provider_user_id TEXT,
    status TEXT NOT NULL DEFAULT 'configured',
    scopes TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create API tokens table (encrypted)
CREATE TABLE IF NOT EXISTS integration.api_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id UUID NOT NULL REFERENCES integration.api_connections(id) ON DELETE CASCADE,
    access_token_encrypted TEXT NOT NULL,
    refresh_token_encrypted TEXT,
    access_token_expires_at TIMESTAMP WITH TIME ZONE,
    refresh_token_expires_at TIMESTAMP WITH TIME ZONE,
    scope TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create channels table for YouTube channels
CREATE TABLE IF NOT EXISTS analytics.channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES core.workspaces(id) ON DELETE CASCADE,
    connection_id UUID NOT NULL REFERENCES integration.api_connections(id) ON DELETE CASCADE,
    platform_id UUID NOT NULL,
    external_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    subscriber_count BIGINT DEFAULT 0,
    view_count BIGINT DEFAULT 0,
    video_count BIGINT DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_api_connections_user_id ON integration.api_connections(user_id);
CREATE INDEX IF NOT EXISTS idx_api_connections_workspace_id ON integration.api_connections(workspace_id);
CREATE INDEX IF NOT EXISTS idx_api_connections_provider_status ON integration.api_connections(provider, status);
CREATE INDEX IF NOT EXISTS idx_api_tokens_connection_id ON integration.api_tokens(connection_id);
CREATE INDEX IF NOT EXISTS idx_channels_workspace_id ON analytics.channels(workspace_id);
CREATE INDEX IF NOT EXISTS idx_channels_connection_id ON analytics.channels(connection_id);
CREATE INDEX IF NOT EXISTS idx_channels_external_id ON analytics.channels(external_id);
CREATE INDEX IF NOT EXISTS idx_channels_status ON analytics.channels(status);

-- Unique constraint
CREATE UNIQUE INDEX IF NOT EXISTS uq_api_connections_user_provider ON integration.api_connections(user_id, provider) WHERE deleted_at IS NULL;
