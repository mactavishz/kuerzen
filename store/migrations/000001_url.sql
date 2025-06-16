-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS urls (
  short_url CHAR(8) PRIMARY KEY,
  long_url VARCHAR(1024) UNIQUE NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- Index for finding existing long URLs (prevent duplicates)
CREATE INDEX idx_urls_long_url ON urls(long_url);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE urls;
-- +goose StatementEnd
