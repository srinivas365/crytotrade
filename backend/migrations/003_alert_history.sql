CREATE TABLE alert_history (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    symbol        TEXT NOT NULL,
    buy_exchange  TEXT NOT NULL,
    sell_exchange TEXT NOT NULL,
    spread_pct    FLOAT NOT NULL,
    buy_price     FLOAT NOT NULL,
    sell_price    FLOAT NOT NULL,
    fired_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX alert_history_user_fired_idx ON alert_history(user_id, fired_at DESC);
