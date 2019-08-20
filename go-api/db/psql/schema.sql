CREATE SCHEMA api;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Works
CREATE TABLE api.works (
  id            uuid PRIMARY KEY DEFAULT uuid_generate_v4(),

  -- basic metadata
  title         TEXT          NOT NULL DEFAULT '',
  description   TEXT          NOT NULL DEFAULT '',
  tags          VARCHAR(64)[] DEFAULT '{}',

  -- extended metadata
  meta          JSONB         NOT NULL DEFAULT '{}'::JSONB,

  created_at    TIMESTAMP     DEFAULT current_timestamp,
  updated_at    TIMESTAMP
);