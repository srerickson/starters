CREATE TABLE works (
  id            INTEGER PRIMARY KEY ASC,
  -- basic metadata
  title         TEXT          NOT NULL DEFAULT '',
  description   TEXT          NOT NULL DEFAULT '',
  -- extended metadata
  meta          JSON          NOT NULL DEFAULT '{}',
  created_at    TIMESTAMP     DEFAULT current_timestamp,
  updated_at    TIMESTAMP
);