ALTER TABLE gjallarhorn.notification ADD COLUMN scheduled_at TIMESTAMPTZ;

-- Partial index over due scheduled rows keeps the dispatcher's claim query fast
-- without bloating the index with already-sent notifications.
CREATE INDEX idx_notification_scheduled ON gjallarhorn.notification (scheduled_at)
    WHERE status = 'SCHEDULED';
