package sqlstore_test

import (
	"testing"

	"github.com/pressly/goose/v3"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/sqlstore"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/storagetest"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

func TestSessionRepositoryContractOnSQLite(t *testing.T) {
	storagetest.RunSessionRepositoryContract(t, func(t *testing.T) ports.SessionRepository {
		db := openSQLite(t)
		if err := sqlstore.Migrate(t.Context(), db, goose.DialectSQLite3); err != nil {
			t.Fatalf("Migrate: %v", err)
		}
		return sqlstore.NewSessionRepository(db)
	})
}
