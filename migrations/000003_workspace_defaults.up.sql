-- Migration: 000003_workspace_defaults
-- Description: Create default roles and update workspace member structure

-- Rename workspace_members to members in core schema (following db design doc)
-- Actually, we'll keep the existing table but rename it

-- Add default roles per workspace
INSERT INTO core.roles (workspace_id, name, description, permissions, is_system)
SELECT 
    w.id,
    role_names.name,
    role_names.description,
    role_names.permissions,
    TRUE
FROM core.workspaces w
CROSS JOIN (
    VALUES 
        ('Owner', 'Full access to all workspace resources', ARRAY['workspace:read','workspace:update','workspace:delete','workspace:create','member:read','member:invite','member:update_role','member:remove','channel:read','channel:create','channel:update','channel:delete','video:read','video:update','analytics:read','audit:read','audit:run','decision:read','decision:manage','task:read','task:manage','report:read','report:generate','settings:read','settings:update','integration:read','integration:manage']),
        ('Admin', 'Manage workspace resources', ARRAY['workspace:read','workspace:update','member:read','member:invite','member:update_role','member:remove','channel:read','channel:create','channel:update','channel:delete','video:read','video:update','analytics:read','audit:read','audit:run','decision:read','decision:manage','task:read','task:manage','report:read','report:generate','settings:read','settings:update','integration:read','integration:manage']),
        ('Member', 'Edit workspace content', ARRAY['workspace:read','channel:read','channel:create','channel:update','video:read','analytics:read','report:read']),
        ('Viewer', 'Read-only access', ARRAY['workspace:read','channel:read','video:read','analytics:read','report:read'])
) AS role_names(name, description, permissions)
WHERE NOT EXISTS (
    SELECT 1 FROM core.roles r 
    WHERE r.workspace_id = w.id AND r.name = role_names.name
);

-- Ensure default roles exist for new workspaces (trigger)
CREATE OR REPLACE FUNCTION core.create_default_workspace_roles()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO core.roles (workspace_id, name, description, permissions, is_system)
    VALUES
        (NEW.id, 'Owner', 'Full access to all workspace resources', 
         ARRAY['workspace:read','workspace:update','workspace:delete','workspace:create','member:read','member:invite','member:update_role','member:remove','channel:read','channel:create','channel:update','channel:delete','video:read','video:update','analytics:read','audit:read','audit:run','decision:read','decision:manage','task:read','task:manage','report:read','report:generate','settings:read','settings:update','integration:read','integration:manage'], TRUE),
        (NEW.id, 'Admin', 'Manage workspace resources', 
         ARRAY['workspace:read','workspace:update','member:read','member:invite','member:update_role','member:remove','channel:read','channel:create','channel:update','channel:delete','video:read','video:update','analytics:read','audit:read','audit:run','decision:read','decision:manage','task:read','task:manage','report:read','report:generate','settings:read','settings:update','integration:read','integration:manage'], TRUE),
        (NEW.id, 'Member', 'Edit workspace content', 
         ARRAY['workspace:read','channel:read','channel:create','channel:update','video:read','analytics:read','report:read'], TRUE),
        (NEW.id, 'Viewer', 'Read-only access', 
         ARRAY['workspace:read','channel:read','video:read','analytics:read','report:read'], TRUE);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_create_default_roles
    AFTER INSERT ON core.workspaces
    FOR EACH ROW
    EXECUTE FUNCTION core.create_default_workspace_roles();
