DROP DATABASE IF EXISTS vigie;

CREATE DATABASE vigie WITH
    ENCODING = 'UTF8'
    CONNECTION LIMIT = -1;

\connect vigie;

DO $$ BEGIN
    CREATE ROLE IF NOT EXISTS vigie WITH LOGIN PASSWORD 'vigie';
EXCEPTION WHEN DUPLICATE_OBJECT THEN
    -- do nothing, role already exists
END $$;

GRANT ALL PRIVILEGES ON DATABASE vigie TO vigie;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO vigie;

COMMENT ON DATABASE vigie IS 'Vigie Dev Database';

-- DROP TABLE IF EXISTS tests;

CREATE TABLE tests(
    id         UUID PRIMARY KEY,
    probe_type VARCHAR(30) NOT NULL,
    interval   INTERVAL    NOT NULL,
    last_run   TIMESTAMP DEFAULT NULL,
    probe_data BYTEA       NOT NULL
);

-- Grant permissions right after creating the table
GRANT ALL ON TABLE public.tests TO vigie;

COMMENT ON COLUMN tests.id IS 'Test ID (test num sha)';
COMMENT ON COLUMN tests.type IS 'Probe type';
COMMENT ON COLUMN tests.frequency IS 'Test Frequency ()';
COMMENT ON COLUMN tests.data IS 'Test Data (protobuf bin)';


CREATE INDEX index_probe_type ON tests ( probe_type );
CREATE INDEX index_id ON tests (  id  );
