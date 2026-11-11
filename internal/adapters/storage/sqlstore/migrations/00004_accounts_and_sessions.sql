-- +goose Up
-- Accounts: someone who has signed in, found by their identity's issuer and
-- subject, and the workspace they work in.
CREATE TABLE accounts (
    id           TEXT PRIMARY KEY NOT NULL,
    issuer       TEXT NOT NULL,
    subject      TEXT NOT NULL,
    name         TEXT NOT NULL,
    workspace_id TEXT NOT NULL
);
CREATE INDEX accounts_by_identity ON accounts (issuer, subject);

-- +goose Down
DROP TABLE accounts;
