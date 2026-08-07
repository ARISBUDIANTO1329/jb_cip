-- Migration: 000001_init_core_schema
-- Description: Create core schema tables for CIP

-- Create schemas
CREATE SCHEMA IF NOT EXISTS core;

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Roles
CREATE TABLE core.roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID NOT NULL,
    name VARCHAR(50) NOT NULL,
    description TEXT,
    permissions TEXT[] DEFAULT '{}',
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Users
CREATE TABLE core.users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    avatar_url VARCHAR(500),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended', 'deleted')),
    email_verified_at TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Workspaces
CREATE TABLE core.workspaces (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id UUID NOT NULL REFERENCES core.users(id),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'archived')),
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Workspace Members
CREATE TABLE core.workspace_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID NOT NULL REFERENCES core.workspaces(id),
    user_id UUID NOT NULL REFERENCES core.users(id),
    role_id UUID NOT NULL REFERENCES core.roles(id),
    invited_by UUID REFERENCES core.users(id),
    invited_at TIMESTAMPTZ DEFAULT NOW(),
    joined_at TIMESTAMPTZ,
    status VARCHAR(20) DEFAULT 'invited' CHECK (status IN ('invited', 'active', 'suspended', 'removed')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(workspace_id, user_id)
);

-- Workspace Settings
CREATE TABLE core.workspace_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID UNIQUE NOT NULL REFERENCES core.workspaces(id),
    key VARCHAR(100) NOT NULL,
    value JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Permissions
CREATE TABLE core.permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Role Permissions (junction)
CREATE TABLE core.role_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    role_id UUID NOT NULL REFERENCES core.roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES core.permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(role_id, permission_id)
);

-- Indexes for performance
CREATE INDEX idx_users_email ON core.users(email);
CREATE INDEX idx_users_status ON core.users(status);
CREATE INDEX idx_users_created_at ON core.users(created_at);

CREATE INDEX idx_workspaces_owner_id ON core.workspaces(owner_id);
CREATE INDEX idx_workspaces_slug ON core.workspaces(slug);
CREATE INDEX idx_workspaces_status ON core.workspaces(status);
CREATE INDEX idx_workspaces_created_at ON core.workspaces(created_at);

CREATE INDEX idx_workspace_members_workspace_id ON core.workspace_members(workspace_id);
CREATE INDEX idx_workspace_members_user_id ON core.workspace_members(user_id);
CREATE INDEX idx_workspace_members_role_id ON core.workspace_members(role_id);
CREATE INDEX idx_workspace_members_status ON core.workspace_members(status);

CREATE INDEX idx_roles_workspace_id ON core.roles(workspace_id);
CREATE INDEX idx_roles_name ON core.roles(name);

CREATE INDEX idx_workspace_settings_workspace_id ON core.workspace_settings(workspace_id);
CREATE INDEX idx_workspace_settings_key ON core.workspace_settings(key);

CREATE INDEX idx_permissions_code ON core.permissions(code);
CREATE INDEX idx_permissions_category ON core.permissions(category);

CREATE INDEX idx_role_permissions_role_id ON core.role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON core.role_permissions(permission_id);

-- Comments
COMMENT ON TABLE core.users IS 'Application users';
COMMENT ON TABLE core.workspaces IS 'Organizational units for channel management';
COMMENT ON TABLE core.workspace_members IS 'User membership in workspaces';
COMMENT ON TABLE core.roles IS 'Role definitions per workspace';
COMMENT ON TABLE core.permissions IS 'Global permission definitions';
COMMENT ON TABLE core.role_permissions IS 'Role-permission mapping';
COMMENT ON TABLE core.workspace_settings IS 'Workspace configuration key-value store';
