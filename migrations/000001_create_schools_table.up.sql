CREATE TABLE IF NOT EXISTS school(
    id bigserial PRIMARY KEY,
    created_at TIMESTAMP (0) with time zone NOT NULL DEFAULT NOW()
);