-- +goose Up
-- Accounts: someone who has signed in, found by their identity's issuer and
-- subject, and the workspace they work in. One identity has at most one
-- account, so the issuer and subject are unique together.
CREATE TABLE accounts (
    id           TEXT PRIMARY KEY NOT NULL,
    issuer       TEXT NOT NULL,
    subject      TEXT NOT NULL,
    name         TEXT NOT NULL,
    workspace_id TEXT NOT NULL
);
CREATE UNIQUE INDEX accounts_by_identity ON accounts (issuer, subject);

-- Sessions: an account's signed-in session, kept by its token's SHA-256 hash
-- only. The expiry is UTC text at full precision, fixed width.
CREATE TABLE sessions (
    token_hash TEXT PRIMARY KEY NOT NULL,
    account_id TEXT NOT NULL,
    expires_at TEXT NOT NULL
);

-- +goose Down
DROP TABLE sessions;
DROP TABLE accounts;
