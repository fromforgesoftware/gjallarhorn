CREATE TABLE gjallarhorn.notification (
    id         UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    realm_id   UUID,
    recipient  TEXT NOT NULL,
    channel    TEXT NOT NULL,
    template   TEXT,
    data       JSONB,
    subject    TEXT,
    body       TEXT,
    status     TEXT NOT NULL,
    last_error TEXT
);

CREATE INDEX idx_notification_status ON gjallarhorn.notification (status);
CREATE INDEX idx_notification_recipient ON gjallarhorn.notification (recipient);

CREATE TABLE gjallarhorn.delivery_attempt (
    id              UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    notification_id UUID NOT NULL REFERENCES gjallarhorn.notification (id) ON DELETE CASCADE,
    attempt         INTEGER NOT NULL,
    status          TEXT NOT NULL,
    error           TEXT,
    attempted_at    TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_delivery_attempt_notification ON gjallarhorn.delivery_attempt (notification_id);
