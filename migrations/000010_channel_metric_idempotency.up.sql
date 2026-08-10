-- Migration: 000010_channel_metric_idempotency
-- Description: Fix channel metric uniqueness using partial unique indexes
-- Root cause: PostgreSQL UNIQUE constraint doesn't prevent duplicates when video_id IS NULL

-- Step 1: Clean up duplicate channel metrics (keep the most recently updated row per group)
DELETE FROM analytics.daily_metrics
WHERE id NOT IN (
    SELECT DISTINCT ON (channel_id, date, metric_type) id
    FROM analytics.daily_metrics
    WHERE video_id IS NULL
    ORDER BY channel_id, date, metric_type, updated_at DESC
)
AND video_id IS NULL;

-- Step 2: Drop the old UNIQUE constraint (which doesn't work with NULL video_id)
ALTER TABLE analytics.daily_metrics
    DROP CONSTRAINT daily_metrics_channel_id_video_id_date_metric_type_key;

-- Step 3: Create partial unique index for VIDEO metrics
CREATE UNIQUE INDEX idx_daily_metrics_video_unique
    ON analytics.daily_metrics(channel_id, video_id, date, metric_type)
    WHERE video_id IS NOT NULL;

-- Step 4: Create partial unique index for CHANNEL metrics
CREATE UNIQUE INDEX idx_daily_metrics_channel_unique
    ON analytics.daily_metrics(channel_id, date, metric_type)
    WHERE video_id IS NULL;
