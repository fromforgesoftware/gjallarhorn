-- notification_preference records per-recipient channel opt-outs. realm_id is
-- TEXT ('' = global) so a nullable realm doesn't break the uniqueness key.
-- A suppressed=TRUE row mutes that channel for the recipient.
CREATE TABLE gjallarhorn.notification_preference (
    id         UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    realm_id   TEXT NOT NULL DEFAULT '',
    recipient  TEXT NOT NULL,
    channel    TEXT NOT NULL,
    suppressed BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE (realm_id, recipient, channel)
);

CREATE INDEX idx_notification_preference_lookup
    ON gjallarhorn.notification_preference (realm_id, recipient, channel);
