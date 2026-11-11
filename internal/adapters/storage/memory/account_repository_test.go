package memory_test

import (
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/memory"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/storagetest"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

func TestAccountRepositoryContract(t *testing.T) {
	storagetest.RunAccountRepositoryContract(t, func(t *testing.T) ports.AccountRepository {
		return memory.NewAccountRepository()
	})
}
