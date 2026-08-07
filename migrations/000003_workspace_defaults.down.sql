-- Rollback: 000003_workspace_defaults

DROP TRIGGER IF EXISTS trg_create_default_roles ON core.workspaces;
DROP FUNCTION IF EXISTS core.create_default_workspace_roles();
DELETE FROM core.roles WHERE is_system = TRUE;
