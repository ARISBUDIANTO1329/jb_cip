-- Migration: 000009_channels_unique_workspace_external (rollback)
DROP INDEX IF EXISTS analytics.uq_channels_workspace_external;
