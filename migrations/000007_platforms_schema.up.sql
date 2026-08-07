-- Migration: 000007_platforms_schema
-- Description: Add analytics.platforms table, seed YouTube platform, link channels FK (07_DATABASE_DESIGN.md 3.5)

CREATE TABLE IF NOT EXISTS analytics.platforms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES core.workspaces(id) ON DELETE CASCADE,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_platforms_workspace_code
    ON analytics.platforms(workspace_id, code) WHERE deleted_at IS NULL;

-- Seed YouTube platform (UUID konsisten dengan channel seed existing: 24803045-...)
INSERT INTO analytics.platforms (id, workspace_id, code, name, description, is_active)
VALUES (
    '24803045-f778-461c-9147-b6fdfc176518',
    '00000000-0000-0000-0000-000000000002',
    'youtube',
    'YouTube',
    'YouTube video platform',
    TRUE
)
ON CONFLICT (workspace_id, code) WHERE deleted_at IS NULL DO NOTHING;

-- Hubungkan channels.platform_id ke analytics.platforms
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_channels_platform') THEN
        ALTER TABLE analytics.channels
            ADD CONSTRAINT fk_channels_platform
            FOREIGN KEY (platform_id) REFERENCES analytics.platforms(id);
    END IF;
END $$;
