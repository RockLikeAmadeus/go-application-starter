package sqlite_test

import (
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/sqlite"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/storagetest"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

func TestWorkspacePlanRepositoryContract(t *testing.T) {
	storagetest.RunWorkspacePlanRepositoryContract(t, func(t *testing.T) ports.WorkspacePlanRepository {
		return sqlite.NewWorkspacePlanRepository(openStore(t, t.TempDir()))
	})
}
