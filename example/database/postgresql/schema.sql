-- Using sqlc examples: https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html

CREATE TABLE authors (
  id   BIGSERIAL PRIMARY KEY,
  name text      NOT NULL,
  bio  text
);
