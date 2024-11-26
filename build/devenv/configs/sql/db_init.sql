DROP DATABASE IF EXISTS vigie;

CREATE DATABASE vigie WITH
    ENCODING = 'UTF8'
    CONNECTION LIMIT = -1;

\connect vigie;

-- Create role if it does not exist
DO $$ BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'vigie') THEN
        CREATE ROLE vigie WITH LOGIN PASSWORD 'vigie';
    END IF;
END $$;

GRANT ALL PRIVILEGES ON DATABASE vigie TO vigie;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO vigie;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO vigie;

COMMENT ON DATABASE vigie IS 'Vigie Dev Database';

-- DROP TABLE IF EXISTS tests;

CREATE TABLE IF NOT EXISTS tests(
    id         UUID PRIMARY KEY,
    probe_type VARCHAR(30) NOT NULL,
    interval   INTERVAL    NOT NULL,
    last_run   TIMESTAMP DEFAULT NULL,
    probe_data BYTEA       NOT NULL
);

-- Grant permissions right after creating the table
GRANT ALL ON TABLE public.tests TO vigie;

-- Add comment to the probe_type column
COMMENT ON COLUMN tests.probe_type IS 'Probe type';
