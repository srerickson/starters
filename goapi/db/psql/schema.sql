CREATE SCHEMA api;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Resources
CREATE TABLE api.resources (
  id            uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  label         TEXT          NOT NULL DEFAULT '',
  fields        JSONB         NOT NULL DEFAULT '{}'::JSONB,
  created_at    TIMESTAMP     DEFAULT current_timestamp,
  updated_at    TIMESTAMP
);