-- Migration: 000005_sync_schema
-- Description: YouTube sync engine tables

-- YouTube Videos table
CREATE TABLE IF NOT EXISTS analytics.youtube_videos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES analytics.channels(id) ON DELETE CASCADE,
    video_id VARCHAR(50) NOT NULL,
    title VARCHAR(500),
    description TEXT,
    published_at TIMESTAMPTZ,
    duration INTEGER, -- in seconds
    thumbnail_url VARCHAR(500),
    privacy_status VARCHAR(20) DEFAULT 'private',
    view_count BIGINT DEFAULT 0,
    like_count BIGINT DEFAULT 0,
    comment_count BIGINT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(channel_id, video_id)
);

-- Sync Jobs table
CREATE TABLE IF NOT EXISTS analytics.sync_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES analytics.channels(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    workspace_id UUID NOT NULL,
    sync_type VARCHAR(20) NOT NULL DEFAULT 'manual', -- initial, manual, incremental
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, running, completed, failed
    total_videos INTEGER DEFAULT 0,
    total_success INTEGER DEFAULT 0,
    total_failed INTEGER DEFAULT 0,
    duration_seconds INTEGER DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_youtube_videos_channel ON analytics.youtube_videos(channel_id);
CREATE INDEX IF NOT EXISTS idx_youtube_videos_published ON analytics.youtube_videos(published_at DESC);
CREATE INDEX IF NOT EXISTS idx_sync_jobs_channel ON analytics.sync_jobs(channel_id);
CREATE INDEX IF NOT EXISTS idx_sync_jobs_status ON analytics.sync_jobs(status);
CREATE INDEX IF NOT EXISTS idx_sync_jobs_created ON analytics.sync_jobs(created_at DESC);

-- Comments
COMMENT ON TABLE analytics.youtube_videos IS 'Synced YouTube videos';
COMMENT ON TABLE analytics.sync_jobs IS 'YouTube sync job history';
