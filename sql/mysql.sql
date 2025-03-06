CREATE TABLE events
(
    id           VARCHAR(255) PRIMARY KEY,
    aggregate_id VARCHAR(255),
    type         VARCHAR(255),
    data         TEXT,
    timestamp    BIGINT
);