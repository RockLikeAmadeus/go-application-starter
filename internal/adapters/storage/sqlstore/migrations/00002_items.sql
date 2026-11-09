-- +goose Up
-- Sample: the Item table. Times are UTC text at full precision, fixed width
-- so they sort in time order.
CREATE TABLE items (
    workspace_id TEXT NOT NULL,
    id           TEXT NOT NULL,
    name         TEXT NOT NULL,
    created_at   TEXT NOT NULL,
    updated_at   TEXT NOT NULL,
    PRIMARY KEY (workspace_id, id)
);

-- +goose Down
DROP TABLE items;
