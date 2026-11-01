// Package storageutil holds helpers shared by storage adapters.
package storageutil

import (
	"cmp"
	"slices"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// WorkspaceItems returns the map's values in workspace, ordered by CreatedAt,
// then ID, which is the contract of ports.ItemRepository.FindAll.
func WorkspaceItems(m map[string]domain.Item, workspace domain.WorkspaceID) []domain.Item {
	items := make([]domain.Item, 0, len(m))
	for _, it := range m {
		if it.WorkspaceID == workspace {
			items = append(items, it)
		}
	}
	slices.SortFunc(items, func(a, b domain.Item) int {
		if c := a.CreatedAt.Compare(b.CreatedAt); c != 0 {
			return c
		}
		return cmp.Compare(a.ID, b.ID)
	})
	return items
}
