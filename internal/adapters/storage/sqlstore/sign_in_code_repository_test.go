package sqlstore_test

import (
	"testing"

	"github.com/pressly/goose/v3"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/sqlstore"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/storagetest"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

func TestSignInCodeRepositoryContractOnSQLite(t *testing.T) {
	storagetest.RunSignInCodeRepositoryContract(t, func(t *testing.T) ports.SignInCodeRepository {
		db := openSQLite(t)
		if err := sqlstore.Migrate(t.Context(), db, goose.DialectSQLite3); err != nil {
			t.Fatalf("Migrate: %v", err)
		}
		return sqlstore.NewSignInCodeRepository(db)
	})
}

func TestSignInCodeRepositoryContractOnPostgres(t *testing.T) {
	storagetest.RunSignInCodeRepositoryContract(t, func(t *testing.T) ports.SignInCodeRepository {
		return sqlstore.NewSignInCodeRepository(openPostgres(t))
	})
}
