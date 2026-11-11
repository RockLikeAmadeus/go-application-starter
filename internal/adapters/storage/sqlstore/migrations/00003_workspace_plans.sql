-- +goose Up
-- Which plan each workspace is on: one row per workspace.
CREATE TABLE workspace_plans (
    workspace_id TEXT PRIMARY KEY NOT NULL,
    plan         TEXT NOT NULL
);

-- +goose Down
DROP TABLE workspace_plans;
