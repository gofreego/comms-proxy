-- In-app notifications: one row per recipient profile, read through the Inbox API.
CREATE TABLE in_app_notifications (
    id UUID PRIMARY KEY,
    app VARCHAR(50) NOT NULL,
    profile_id BIGINT NOT NULL,
    category VARCHAR(100) NOT NULL,
    title VARCHAR(200) NOT NULL,
    body TEXT NOT NULL DEFAULT '',
    link TEXT NOT NULL DEFAULT '',
    image_url TEXT NOT NULL DEFAULT '',
    data JSONB NOT NULL DEFAULT '{}',
    dedup_key VARCHAR(200) NOT NULL DEFAULT '',
    created_at BIGINT NOT NULL,
    read_at BIGINT NOT NULL DEFAULT 0
);

-- A profile's inbox, newest first.
CREATE INDEX idx_in_app_notifications_inbox ON in_app_notifications (app, profile_id, created_at DESC);
-- Unread counts (the app's badge) touch only unread rows.
CREATE INDEX idx_in_app_notifications_unread ON in_app_notifications (app, profile_id) WHERE read_at = 0;
-- A dedup key is sent to each profile at most once.
CREATE UNIQUE INDEX idx_in_app_notifications_dedup ON in_app_notifications (app, profile_id, dedup_key) WHERE dedup_key <> '';
