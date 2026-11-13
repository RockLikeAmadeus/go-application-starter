package memory_test

import (
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/memory"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/storagetest"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

func TestSignInCodeRepositoryContract(t *testing.T) {
	storagetest.RunSignInCodeRepositoryContract(t, func(t *testing.T) ports.SignInCodeRepository {
		return memory.NewSignInCodeRepository()
	})
}
