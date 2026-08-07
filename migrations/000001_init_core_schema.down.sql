-- Rollback: 000001_init_core_schema

DROP TABLE IF EXISTS core.role_permissions;
DROP TABLE IF EXISTS core.workspace_settings;
DROP TABLE IF EXISTS core.workspace_members;
DROP TABLE IF EXISTS core.workspaces;
DROP TABLE IF EXISTS core.users;
DROP TABLE IF EXISTS core.roles;
DROP TABLE IF EXISTS core.permissions;

DROP SCHEMA IF EXISTS core;
