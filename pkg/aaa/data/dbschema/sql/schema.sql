-- Version: 0.0
-- Description: Create table test
CREATE TABLE tests (
                       id INT NOT NULL,
                       probe_type VARCHAR(30) NOT NULL,
                       frequency INT NOT NULL,
                       last_run INT DEFAULT NULL,
                       probe_data BYTEA NOT NULL
);

