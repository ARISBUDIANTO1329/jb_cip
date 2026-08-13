-- Migration: 000011_audit_snapshots
-- Description: Store weekly audit snapshots for comparison

CREATE TABLE IF NOT EXISTS analytics.audit_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES analytics.channels(id) ON DELETE CASCADE,
    snapshot_date DATE NOT NULL,
    week_number INTEGER NOT NULL,
    year INTEGER NOT NULL,
    
    findings JSONB NOT NULL DEFAULT '[]',
    summary JSONB NOT NULL DEFAULT '{}',
    
    total_findings INTEGER DEFAULT 0,
    critical_count INTEGER DEFAULT 0,
    high_count INTEGER DEFAULT 0,
    medium_count INTEGER DEFAULT 0,
    low_count INTEGER DEFAULT 0,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(channel_id, year, week_number)
);

CREATE INDEX IF NOT EXISTS idx_audit_snapshots_channel ON analytics.audit_snapshots(channel_id);
CREATE INDEX IF NOT EXISTS idx_audit_snapshots_date ON analytics.audit_snapshots(snapshot_date DESC);

COMMENT ON TABLE analytics.audit_snapshots IS 'Weekly audit snapshots for tracking improvements over time';
