package sqlstore_test

import (
	"testing"

	"github.com/pressly/goose/v3"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/sqlstore"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/storagetest"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

func TestAccountRepositoryContractOnSQLite(t *testing.T) {
	storagetest.RunAccountRepositoryContract(t, func(t *testing.T) ports.AccountRepository {
		db := openSQLite(t)
		if err := sqlstore.Migrate(t.Context(), db, goose.DialectSQLite3); err != nil {
			t.Fatalf("Migrate: %v", err)
		}
		return sqlstore.NewAccountRepository(db)
	})
}

func TestAccountRepositoryContractOnPostgres(t *testing.T) {
	storagetest.RunAccountRepositoryContract(t, func(t *testing.T) ports.AccountRepository {
		return sqlstore.NewAccountRepository(openPostgres(t))
	})
}
