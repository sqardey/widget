CREATE TABLE IF NOT EXISTS widget_users (
    id BIGSERIAL PRIMARY KEY,
    tg_id BIGINT UNIQUE,
    key TEXT NOT NULL UNIQUE DEFAULT gen_random_uuid()::text,
    login TEXT NOT NULL,
    password TEXT NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_widget_users_key ON widget_users(key);
CREATE INDEX IF NOT EXISTS idx_widget_users_tg_id ON widget_users(tg_id);