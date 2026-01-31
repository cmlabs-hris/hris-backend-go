-- =========================
-- Notifications Migration Down
-- =========================

-- Drop indexes
DROP INDEX IF EXISTS idx_notification_preferences_user;
DROP INDEX IF EXISTS idx_notifications_company;
DROP INDEX IF EXISTS idx_notifications_recipient;

-- Drop tables in reverse order (notification_preferences has FK to users, notifications has FK to users)
DROP TABLE IF EXISTS notification_preferences;
DROP TABLE IF EXISTS notifications;
