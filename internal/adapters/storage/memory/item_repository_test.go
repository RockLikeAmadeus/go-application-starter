package memory_test

import (
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/memory"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/storagetest"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

func TestItemRepositoryContract(t *testing.T) {
	storagetest.RunItemRepositoryContract(t, func(t *testing.T) ports.ItemRepository {
		return memory.NewItemRepository()
	})
}
