package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// timeFormat keeps full precision and a fixed width, so stored times sort in
// time order as text.
const timeFormat = "2006-01-02T15:04:05.000000000Z"

const itemColumns = `workspace_id, id, name, created_at, updated_at`

// ItemRepository persists Items in the store's items table.
type ItemRepository struct {
	store *Store
}

func NewItemRepository(store *Store) *ItemRepository {
	return &ItemRepository{store: store}
}

func (r *ItemRepository) Save(ctx context.Context, item domain.Item) error {
	_, err := r.store.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO items (`+itemColumns+`) VALUES (?, ?, ?, ?, ?)`,
		item.WorkspaceID, item.ID, item.Name, formatTime(item.CreatedAt), formatTime(item.UpdatedAt))
	return err
}

func (r *ItemRepository) FindByID(ctx context.Context, workspace domain.WorkspaceID, id string) (domain.Item, error) {
	row := r.store.db.QueryRowContext(ctx,
		`SELECT `+itemColumns+` FROM items WHERE workspace_id = ? AND id = ?`, workspace, id)
	item, err := scanItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Item{}, domain.ErrNotFound
	}
	return item, err
}

func (r *ItemRepository) FindAll(ctx context.Context, workspace domain.WorkspaceID) ([]domain.Item, error) {
	rows, err := r.store.db.QueryContext(ctx,
		`SELECT `+itemColumns+` FROM items WHERE workspace_id = ? ORDER BY created_at, id`, workspace)
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
	res, err := r.store.db.ExecContext(ctx,
		`DELETE FROM items WHERE workspace_id = ? AND id = ?`, workspace, id)
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
	if item.CreatedAt, err = time.Parse(timeFormat, created); err != nil {
		return domain.Item{}, err
	}
	if item.UpdatedAt, err = time.Parse(timeFormat, updated); err != nil {
		return domain.Item{}, err
	}
	return item, nil
}

func formatTime(t time.Time) string {
	return t.UTC().Format(timeFormat)
}
