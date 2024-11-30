CREATE TABLE IF NOT EXISTS urls (
  id UUID PRIMARY KEY,
  short_url TEXT UNIQUE NOT NULL,
  original_url TEXT UNIQUE NOT NULL
);

CREATE INDEX idx_short_url ON urls(short_url);

CREATE INDEX idx_original_url ON urls(original_url);