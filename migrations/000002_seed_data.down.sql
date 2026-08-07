-- Rollback: 000002_seed_data

DELETE FROM core.role_permissions;
DELETE FROM core.permissions;
DELETE FROM core.workspace_members;
DELETE FROM core.workspaces;
DELETE FROM core.users;
