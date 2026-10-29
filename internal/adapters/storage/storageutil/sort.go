// Package storageutil holds helpers shared by storage adapters.
package storageutil

import (
	"cmp"
	"slices"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// SortedItems returns the map's values ordered by CreatedAt, then ID, which
// is the ordering contract of ports.ItemRepository.FindAll.
func SortedItems(m map[string]domain.Item) []domain.Item {
	items := make([]domain.Item, 0, len(m))
	for _, it := range m {
		items = append(items, it)
	}
	slices.SortFunc(items, func(a, b domain.Item) int {
		if c := a.CreatedAt.Compare(b.CreatedAt); c != 0 {
			return c
		}
		return cmp.Compare(a.ID, b.ID)
	})
	return items
}
