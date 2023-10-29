
DROP DATABASE IF EXISTS vigie;

CREATE DATABASE vigie WITH
    ENCODING = 'UTF8'
    CONNECTION LIMIT = -1;

COMMENT ON DATABASE vigie IS 'Vigie Dev Database';

\connect vigie

-- DROP TABLE IF EXISTS tests;

CREATE TABLE tests(
    id         UUID PRIMARY KEY,
    probe_type VARCHAR(30) NOT NULL,
    interval   INTERVAL    NOT NULL,
    last_run   TIMESTAMP DEFAULT NULL,
    probe_data BYTEA       NOT NULL
);
/*
COMMENT ON COLUMN tests.id IS 'Test ID (test num sha)';
COMMENT ON COLUMN tests.type IS 'Probe type';
COMMENT ON COLUMN tests.frequency IS 'Test Frequency ()';
COMMENT ON COLUMN tests.data IS 'Test Data (protobuf bin)';

 */

CREATE INDEX index_probe_type ON tests ( probe_type );
CREATE INDEX index_id ON tests (  id  );