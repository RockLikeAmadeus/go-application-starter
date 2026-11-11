package sqlstore

import (
	"context"
	"database/sql"
	"errors"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

const itemColumns = `workspace_id, id, name, created_at, updated_at`

// ItemRepository persists Items in the items table.
type ItemRepository struct {
	db *sql.DB
}

func NewItemRepository(db *sql.DB) *ItemRepository {
	return &ItemRepository{db: db}
}

func (r *ItemRepository) Save(ctx context.Context, item domain.Item) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO items (`+itemColumns+`) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (workspace_id, id) DO UPDATE SET
			name = excluded.name, created_at = excluded.created_at, updated_at = excluded.updated_at`,
		item.WorkspaceID, item.ID, item.Name, formatTime(item.CreatedAt), formatTime(item.UpdatedAt))
	return err
}

func (r *ItemRepository) FindByID(ctx context.Context, workspace domain.WorkspaceID, id string) (domain.Item, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+itemColumns+` FROM items WHERE workspace_id = $1 AND id = $2`, workspace, id)
	item, err := scanItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Item{}, domain.ErrNotFound
	}
	return item, err
}

func (r *ItemRepository) FindAll(ctx context.Context, workspace domain.WorkspaceID) ([]domain.Item, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+itemColumns+` FROM items WHERE workspace_id = $1 ORDER BY created_at, id`, workspace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Item{}
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *ItemRepository) Delete(ctx context.Context, workspace domain.WorkspaceID, id string) error {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM items WHERE workspace_id = $1 AND id = $2`, workspace, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// scanItem reads one row of itemColumns.
func scanItem(row interface{ Scan(...any) error }) (domain.Item, error) {
	var item domain.Item
	var created, updated string
	if err := row.Scan(&item.WorkspaceID, &item.ID, &item.Name, &created, &updated); err != nil {
		return domain.Item{}, err
	}
	var err error
	if item.CreatedAt, err = parseTime(created); err != nil {
		return domain.Item{}, err
	}
	if item.UpdatedAt, err = parseTime(updated); err != nil {
		return domain.Item{}, err
	}
	return item, nil
}
