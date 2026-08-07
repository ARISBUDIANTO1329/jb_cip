-- Migration: 000002_seed_data
-- Description: Seed initial data for CIP - Super admin user and default permissions

-- Insert Super Admin user (password: admin123)
INSERT INTO core.users (id, email, password_hash, name, status, email_verified_at)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'admin@cip.local',
    '$2a$12$wGpCl30tZJssYgYDXL20veNTExc/dhbWHycfRpa5vGp6TOO2UyNie',
    'Super Admin',
    'active',
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- Insert default workspace for Super Admin
INSERT INTO core.workspaces (id, owner_id, name, slug, description, status)
VALUES (
    '00000000-0000-0000-0000-000000000002',
    '00000000-0000-0000-0000-000000000001',
    'Default Workspace',
    'default',
    'Default workspace for Super Admin',
    'active'
) ON CONFLICT (slug) DO NOTHING;

-- Insert default permissions
INSERT INTO core.permissions (code, name, description, category) VALUES
    ('workspace:read', 'Read Workspace', 'View workspace details', 'workspace'),
    ('workspace:update', 'Update Workspace', 'Update workspace settings', 'workspace'),
    ('workspace:delete', 'Delete Workspace', 'Delete workspace', 'workspace'),
    ('workspace:create', 'Create Workspace', 'Create new workspace', 'workspace'),
    ('member:read', 'Read Members', 'View workspace members', 'member'),
    ('member:invite', 'Invite Member', 'Invite new members', 'member'),
    ('member:update_role', 'Update Member Role', 'Change member roles', 'member'),
    ('member:remove', 'Remove Member', 'Remove members from workspace', 'member'),
    ('channel:read', 'Read Channels', 'View channels', 'channel'),
    ('channel:create', 'Create Channel', 'Add new channels', 'channel'),
    ('channel:update', 'Update Channel', 'Update channel settings', 'channel'),
    ('channel:delete', 'Delete Channel', 'Remove channels', 'channel'),
    ('video:read', 'Read Videos', 'View videos', 'video'),
    ('video:update', 'Update Video', 'Update video metadata', 'video'),
    ('analytics:read', 'Read Analytics', 'View analytics data', 'analytics'),
    ('audit:read', 'Read Audit', 'View audit results', 'audit'),
    ('audit:run', 'Run Audit', 'Trigger audits', 'audit'),
    ('decision:read', 'Read Decisions', 'View decisions', 'decision'),
    ('decision:manage', 'Manage Decisions', 'Accept/reject decisions', 'decision'),
    ('task:read', 'Read Tasks', 'View tasks', 'task'),
    ('task:manage', 'Manage Tasks', 'Create/complete tasks', 'task'),
    ('report:read', 'Read Reports', 'View reports', 'report'),
    ('report:generate', 'Generate Report', 'Create reports', 'report'),
    ('settings:read', 'Read Settings', 'View settings', 'settings'),
    ('settings:update', 'Update Settings', 'Update settings', 'settings'),
    ('integration:read', 'Read Integrations', 'View integrations', 'integration'),
    ('integration:manage', 'Manage Integrations', 'Connect/disconnect integrations', 'integration') ON CONFLICT (code) DO NOTHING;

COMMENT ON TABLE core.users IS 'Application users';
COMMENT ON TABLE core.permissions IS 'Global permission definitions for CIP';
