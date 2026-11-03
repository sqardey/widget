CREATE TABLE IF NOT EXISTS widget_users(
    id BIGSERIAL PRIMARY KEY,
    key TEXT NOT NULL,
    login TEXT NOT NULL,
    password TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_widget_users_key ON widget_users(key);