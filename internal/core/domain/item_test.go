package domain_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

var now = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

const ws domain.WorkspaceID = "ws-1"

func TestNewItem_SetsWorkspace(t *testing.T) {
	item, err := domain.NewItem(ws, "id-1", "Widget", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.WorkspaceID != ws {
		t.Errorf("WorkspaceID = %q, want %q", item.WorkspaceID, ws)
	}
}

func TestNewItem_TrimsName(t *testing.T) {
	item, err := domain.NewItem(ws, "id-1", "  Widget  ", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Name != "Widget" {
		t.Errorf("Name = %q, want %q", item.Name, "Widget")
	}
	if !item.CreatedAt.Equal(now) || !item.UpdatedAt.Equal(now) {
		t.Errorf("timestamps not set to now")
	}
}

func TestNewItem_RejectsInvalidNames(t *testing.T) {
	cases := map[string]string{
		"empty":      "",
		"whitespace": "   ",
		"too long":   strings.Repeat("x", domain.MaxItemNameLength+1),
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := domain.NewItem(ws, "id-1", input, now)
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("err = %v, want ErrValidation", err)
			}
		})
	}
}

func TestItem_Rename(t *testing.T) {
	item, _ := domain.NewItem(ws, "id-1", "Old", now)
	later := now.Add(time.Hour)

	if err := item.Rename("New", later); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Name != "New" || !item.UpdatedAt.Equal(later) || !item.CreatedAt.Equal(now) {
		t.Errorf("unexpected item after rename: %+v", item)
	}
}

func TestItem_RenameRejectsEmpty(t *testing.T) {
	item, _ := domain.NewItem(ws, "id-1", "Old", now)
	if err := item.Rename(" ", now); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("err = %v, want ErrValidation", err)
	}
	if item.Name != "Old" {
		t.Errorf("name changed on failed rename: %q", item.Name)
	}
}

func TestItemsCSV(t *testing.T) {
	first, _ := domain.NewItem(ws, "id-1", "Plain", now)
	second, _ := domain.NewItem(ws, "id-2", `Comma, and "quotes"`, now.Add(time.Hour))

	got := string(domain.ItemsCSV([]domain.Item{first, second}))

	want := "name,createdAt\n" +
		"Plain,2026-01-02T03:04:05Z\n" +
		`"Comma, and ""quotes""",2026-01-02T04:04:05Z` + "\n"
	if got != want {
		t.Errorf("ItemsCSV =\n%s\nwant\n%s", got, want)
	}
}
