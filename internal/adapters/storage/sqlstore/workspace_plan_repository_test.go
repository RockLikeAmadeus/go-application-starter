package sqlstore_test

import (
	"testing"

	"github.com/pressly/goose/v3"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/sqlstore"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/storagetest"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

func TestWorkspacePlanRepositoryContractOnSQLite(t *testing.T) {
	storagetest.RunWorkspacePlanRepositoryContract(t, func(t *testing.T) ports.WorkspacePlanRepository {
		db := openSQLite(t)
		if err := sqlstore.Migrate(t.Context(), db, goose.DialectSQLite3); err != nil {
			t.Fatalf("Migrate: %v", err)
		}
		return sqlstore.NewWorkspacePlanRepository(db)
	})
}
