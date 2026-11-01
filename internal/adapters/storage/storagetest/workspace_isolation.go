// Package storagetest holds the contract test suites that storage adapters
// must pass. A new adapter should start by calling them from its own tests.
package storagetest

import (
	"context"
	"errors"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// Records tells RunWorkspaceIsolation how to make and read the records of one
// entity.
type Records[T any] struct {
	// New returns a valid record with the given workspace and ID.
	New func(workspace domain.WorkspaceID, id string) T
	// ID and Workspace read a record's ID and workspace.
	ID        func(T) string
	Workspace func(T) domain.WorkspaceID
}

// RunWorkspaceIsolation checks that a repository keeps workspaces apart: a
// record is only seen, and only deleted, in its own workspace. Every
// ports.WorkspaceRepository implementation must pass it. newRepo returns a
// fresh, empty repository for each subtest.
func RunWorkspaceIsolation[T any](t *testing.T, newRepo func(*testing.T) ports.WorkspaceRepository[T], records Records[T]) {
	ctx := context.Background()
	const mine, theirs domain.WorkspaceID = "ws-a", "ws-b"

	save := func(t *testing.T, repo ports.WorkspaceRepository[T], record T) {
		t.Helper()
		if err := repo.Save(ctx, record); err != nil {
			t.Fatalf("Save(%s): %v", records.ID(record), err)
		}
	}

	t.Run("Save keeps the record's workspace", func(t *testing.T) {
		repo := newRepo(t)
		save(t, repo, records.New(theirs, "a"))

		got, err := repo.FindByID(ctx, theirs, "a")
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if records.Workspace(got) != theirs {
			t.Errorf("workspace = %q, want %q", records.Workspace(got), theirs)
		}
	})

	t.Run("FindAll lists only the given workspace's records", func(t *testing.T) {
		repo := newRepo(t)
		save(t, repo, records.New(mine, "a"))
		save(t, repo, records.New(theirs, "b"))

		all, err := repo.FindAll(ctx, mine)
		if err != nil {
			t.Fatalf("FindAll: %v", err)
		}
		if len(all) != 1 || records.ID(all[0]) != "a" {
			var ids []string
			for _, r := range all {
				ids = append(ids, records.ID(r))
			}
			t.Errorf("FindAll(%s) IDs = %v, want only [a]", mine, ids)
		}
	})

	t.Run("FindByID of another workspace's record returns ErrNotFound", func(t *testing.T) {
		repo := newRepo(t)
		save(t, repo, records.New(theirs, "a"))

		if _, err := repo.FindByID(ctx, mine, "a"); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("Delete of another workspace's record returns ErrNotFound and keeps it", func(t *testing.T) {
		repo := newRepo(t)
		save(t, repo, records.New(theirs, "a"))

		if err := repo.Delete(ctx, mine, "a"); !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("err = %v, want ErrNotFound", err)
		}
		if _, err := repo.FindByID(ctx, theirs, "a"); err != nil {
			t.Errorf("FindByID in its own workspace after the delete: %v", err)
		}
	})
}
