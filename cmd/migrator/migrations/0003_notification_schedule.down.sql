DROP INDEX IF EXISTS gjallarhorn.idx_notification_scheduled;
ALTER TABLE gjallarhorn.notification DROP COLUMN IF EXISTS scheduled_at;
