-- Migration: 000007_platforms_schema (rollback)
ALTER TABLE analytics.channels DROP CONSTRAINT IF EXISTS fk_channels_platform;
DROP TABLE IF EXISTS analytics.platforms;
