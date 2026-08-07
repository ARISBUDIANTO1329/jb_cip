-- Migration: 000006_daily_metrics_schema
-- Description: Daily metrics table for YouTube Analytics

-- Daily Metrics table
CREATE TABLE IF NOT EXISTS analytics.daily_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES analytics.channels(id) ON DELETE CASCADE,
    video_id UUID REFERENCES analytics.youtube_videos(id) ON DELETE SET NULL,
    date DATE NOT NULL,
    metric_type VARCHAR(50) NOT NULL, -- 'channel', 'video'
    
    -- Video metrics
    views BIGINT DEFAULT 0,
    watch_time BIGINT DEFAULT 0, -- in seconds
    average_view_duration DECIMAL(10,2) DEFAULT 0,
    average_percentage_viewed DECIMAL(5,2) DEFAULT 0,
    impressions BIGINT DEFAULT 0,
    impression_ctr DECIMAL(10,4) DEFAULT 0,
    likes BIGINT DEFAULT 0,
    comments BIGINT DEFAULT 0,
    shares BIGINT DEFAULT 0,
    
    -- Channel metrics
    subscribers BIGINT DEFAULT 0,
    returning_viewers BIGINT DEFAULT 0,
    new_viewers BIGINT DEFAULT 0,
    
    -- Sync metadata
    sync_job_id UUID REFERENCES analytics.sync_jobs(id) ON DELETE SET NULL,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    sync_duration INTEGER DEFAULT 0,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(channel_id, video_id, date, metric_type)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_daily_metrics_channel ON analytics.daily_metrics(channel_id);
CREATE INDEX IF NOT EXISTS idx_daily_metrics_video ON analytics.daily_metrics(video_id);
CREATE INDEX IF NOT EXISTS idx_daily_metrics_date ON analytics.daily_metrics(date);
CREATE INDEX IF NOT EXISTS idx_daily_metrics_sync_job ON analytics.daily_metrics(sync_job_id);

-- Comments
COMMENT ON TABLE analytics.daily_metrics IS 'Daily metrics for YouTube channel and video analytics';
