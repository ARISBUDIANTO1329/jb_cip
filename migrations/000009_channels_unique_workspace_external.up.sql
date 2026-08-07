-- Migration: 000009_channels_unique_workspace_external
-- Description: Fix missing unique constraint on channels(workspace_id, external_id) required by ON CONFLICT in SaveChannels

CREATE UNIQUE INDEX IF NOT EXISTS uq_channels_workspace_external
    ON analytics.channels(workspace_id, external_id) WHERE deleted_at IS NULL;
