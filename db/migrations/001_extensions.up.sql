-- Enable extensions used by the schema
CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Keep timestamps fresh
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS trigger AS $$
BEGIN
  NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;
