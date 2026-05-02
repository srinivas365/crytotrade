CREATE TABLE user_settings (
    user_id           UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    threshold_pct     FLOAT NOT NULL DEFAULT 0.1,
    telegram_bot_token TEXT NOT NULL DEFAULT '',
    telegram_chat_id  TEXT NOT NULL DEFAULT '',
    in_app_alerts     BOOL NOT NULL DEFAULT true,
    alert_sound       BOOL NOT NULL DEFAULT true
);
