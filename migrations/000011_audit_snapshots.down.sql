-- Migration: 000011_audit_snapshots (DOWN)
-- Description: Drop audit snapshots table

DROP TABLE IF EXISTS analytics.audit_snapshots;
