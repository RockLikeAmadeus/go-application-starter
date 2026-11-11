package memory_test

import (
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/memory"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/storagetest"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

func TestSessionRepositoryContract(t *testing.T) {
	storagetest.RunSessionRepositoryContract(t, func(t *testing.T) ports.SessionRepository {
		return memory.NewSessionRepository()
	})
}
