-- Per-realm, localizable notification templates (Go text/template). Send may
-- reference a template by name + locale; the usecase renders subject/body from
-- the notification's data before dispatch. locale '' is the default/fallback.

CREATE TABLE gjallarhorn.notification_template (
  id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  realm_id   UUID NOT NULL,
  name       TEXT NOT NULL,
  channel    TEXT NOT NULL DEFAULT 'EMAIL',
  locale     TEXT NOT NULL DEFAULT '',
  subject    TEXT NOT NULL DEFAULT '',
  body       TEXT NOT NULL DEFAULT '',
  UNIQUE (realm_id, name, locale)
);
CREATE INDEX idx_notification_template_lookup ON gjallarhorn.notification_template (realm_id, name, locale);

ALTER TABLE gjallarhorn.notification ADD COLUMN locale TEXT NOT NULL DEFAULT '';
