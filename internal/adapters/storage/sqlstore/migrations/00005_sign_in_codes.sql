-- +goose Up
-- Sign-in codes: a one-time code that a browser sign-in hands to the command
-- line, kept by its SHA-256 hash only, with the account that signed in, the
-- PKCE challenge it is bound to, and its expiry (UTC text at full precision,
-- fixed width).
CREATE TABLE sign_in_codes (
    code_hash  TEXT PRIMARY KEY NOT NULL,
    account_id TEXT NOT NULL,
    challenge  TEXT NOT NULL,
    expires_at TEXT NOT NULL
);

-- +goose Down
DROP TABLE sign_in_codes;
