-- Migration: 000010_channel_metric_idempotency (DOWN)
-- Description: Revert channel metric idempotency fix

-- Drop partial unique indexes
DROP INDEX IF EXISTS analytics.idx_daily_metrics_video_unique;
DROP INDEX IF EXISTS analytics.idx_daily_metrics_channel_unique;

-- Restore the original UNIQUE constraint
ALTER TABLE analytics.daily_metrics
    ADD CONSTRAINT daily_metrics_channel_id_video_id_date_metric_type_key
    UNIQUE (channel_id, video_id, date, metric_type);
