package bootstrap_test

import (
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/app"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/testsupport"
)

// The local edition's one workspace is on the local plan.
func TestNewEmbedded_IsOnTheLocalPlan(t *testing.T) {
	a, err := bootstrap.NewEmbedded(config.Config{Storage: config.StorageMemory})
	if err != nil {
		t.Fatalf("NewEmbedded: %v", err)
	}
	t.Cleanup(func() { _ = a.Close() })
	if a.Entitlements == nil {
		t.Fatal("the App has no Entitlements")
	}

	got, err := a.Entitlements.CurrentEntitlements(context.Background())
	if err != nil {
		t.Fatalf("CurrentEntitlements: %v", err)
	}
	if got.Plan != "local" || len(got.Limits) != 0 {
		t.Errorf("CurrentEntitlements = %+v, want the local plan with no limits", got)
	}
}

// Storage on disk is ready only while its data folder can be used.
func TestNewEmbedded_IsNotReadyOnceTheDataFolderIsRemoved(t *testing.T) {
	for _, storage := range []config.Storage{config.StorageJSONFile, config.StorageSQLite} {
		t.Run(string(storage), func(t *testing.T) {
			dir := t.TempDir()
			a, err := bootstrap.NewEmbedded(config.Config{Storage: storage, DataDir: dir})
			if err != nil {
				t.Fatalf("NewEmbedded: %v", err)
			}
			t.Cleanup(func() { _ = a.Close() })
			if err := a.Ready(t.Context()); err != nil {
				t.Fatalf("Ready error = %v before removal, want nil", err)
			}
			if err := os.RemoveAll(dir); err != nil {
				t.Fatal(err)
			}

			if err := a.Ready(t.Context()); err == nil {
				t.Error("Ready error = nil, want an error once the data folder is gone")
			}
		})
	}
}

// A build whose local plan isn't in the plans file fails at start, naming
// the plan, rather than on the first request.
func TestNewEmbedded_RefusesALocalPlanThePlansFileDoesntHave(t *testing.T) {
	defer func(plan string) { bootstrap.LocalPlan = plan }(bootstrap.LocalPlan)
	bootstrap.LocalPlan = "gold"

	a, err := bootstrap.NewEmbedded(config.Config{Storage: config.StorageMemory})
	if err == nil {
		_ = a.Close()
		t.Fatal("NewEmbedded succeeded, want an error naming the gold plan")
	}
	if !strings.Contains(err.Error(), `"gold"`) {
		t.Errorf("err = %v, want it to name the gold plan", err)
	}
}

func TestNewEmbedded_WithAPlansFile(t *testing.T) {
	t.Run("uses it instead of the shipped one", func(t *testing.T) {
		path := writePlansFile(t, `{"plans": {"local": {"limits": {"items": {"max": 5, "whenOver": "read-only"}}}}}`)
		a, err := bootstrap.NewEmbedded(config.Config{Storage: config.StorageMemory, PlansFile: path})
		if err != nil {
			t.Fatalf("NewEmbedded: %v", err)
		}
		t.Cleanup(func() { _ = a.Close() })

		got, err := a.Entitlements.CurrentEntitlements(context.Background())
		if err != nil {
			t.Fatalf("CurrentEntitlements: %v", err)
		}
		want := domain.Limit{Name: "items", Max: 5, WhenOver: domain.ReadOnly}
		if got.Limits["items"] != want {
			t.Errorf("items limit = %+v, want %+v from the file", got.Limits["items"], want)
		}
	})
	t.Run("fails at start when the file is bad", func(t *testing.T) {
		path := writePlansFile(t, `{"plans": `)
		a, err := bootstrap.NewEmbedded(config.Config{Storage: config.StorageMemory, PlansFile: path})
		if err == nil {
			_ = a.Close()
			t.Fatal("NewEmbedded succeeded, want the plans file's error")
		}
		if !strings.Contains(err.Error(), "plans file") {
			t.Errorf("err = %v, want it to be about the plans file", err)
		}
	})
}

// writePlansFile writes a plans file into a temporary folder and returns its
// path.
func writePlansFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "plans.jsonc")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// With --server, the CLI and TUI see the plan of the workspace the server
// puts them in.
func TestNewRemote_IsOnTheServersPlan(t *testing.T) {
	pro := domain.Entitlements{Plan: "pro", Limits: map[string]domain.Limit{}}
	core := app.NewEntitlementsService(testsupport.FixedEntitlements(pro), testsupport.FixedWorkspace("ws-a"))
	srv := httptest.NewServer(httpapi.NewHandler(httpapi.Options{Entitlements: core}))
	t.Cleanup(srv.Close)

	a, err := bootstrap.NewRemote(srv.URL)
	if err != nil {
		t.Fatalf("NewRemote: %v", err)
	}
	if a.Entitlements == nil {
		t.Fatal("the App has no Entitlements")
	}
	got, err := a.Entitlements.CurrentEntitlements(context.Background())
	if err != nil {
		t.Fatalf("CurrentEntitlements: %v", err)
	}
	if got.Plan != "pro" {
		t.Errorf("plan = %q, want the server's %q", got.Plan, "pro")
	}
}
