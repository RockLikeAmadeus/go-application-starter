// Package domain holds the entities and business rules of the application.
// It must not import anything outside the standard library.
package domain

import (
	"bytes"
	"encoding/csv"
	"strings"
	"time"
	"unicode/utf8"
)

// MaxItemNameLength is the maximum number of characters allowed in an item name.
const MaxItemNameLength = 200

// Item is the sample aggregate for the walking skeleton. Rename or replace it
// when starting a real project.
type Item struct {
	ID          string
	WorkspaceID WorkspaceID
	Name        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewItem builds a valid Item or returns a *ValidationError.
func NewItem(workspace WorkspaceID, id, name string, now time.Time) (Item, error) {
	name, err := normalizeName(name)
	if err != nil {
		return Item{}, err
	}
	return Item{
		ID:          id,
		WorkspaceID: workspace,
		Name:        name,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Rename changes the item's name, enforcing the same rules as NewItem.
func (i *Item) Rename(name string, now time.Time) error {
	name, err := normalizeName(name)
	if err != nil {
		return err
	}
	i.Name = name
	i.UpdatedAt = now
	return nil
}

func normalizeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", &ValidationError{Field: "name", Message: "name must not be empty"}
	}
	if utf8.RuneCountInString(name) > MaxItemNameLength {
		return "", &ValidationError{Field: "name", Message: "name must be at most 200 characters"}
	}
	return name, nil
}

// ItemsCSV encodes items as CSV, in the order given: a header row of
// name,createdAt, then one row per item, its creation time in RFC 3339.
// Writing to a bytes.Buffer can't fail, so the writer's errors are ignored.
func ItemsCSV(items []Item) []byte {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"name", "createdAt"})
	for _, item := range items {
		_ = w.Write([]string{item.Name, item.CreatedAt.Format(time.RFC3339)})
	}
	w.Flush()
	return buf.Bytes()
}
