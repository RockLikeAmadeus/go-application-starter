package bootstrap_test

import (
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
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
	for _, storage := range []config.StorageBackend{config.StorageJSONFile, config.StorageSQLite} {
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

// The CLI's help reads only the local plan's entitlements: it opens no
// storage, so it creates no data folder and isn't stopped by a locked one.
// The local edition gives no upgrade hints, so a locked feature names no
// plan.
func TestNewEntitlements_ReadsTheLocalPlanWithoutStorageOrHints(t *testing.T) {
	path := writePlansFile(t, `{
		"features": {"export": {"name": "Export", "whenOff": "show-locked"}},
		"plans": {"local": {"limits": {}}, "pro": {"limits": {}, "features": ["export"]}}
	}`)
	dataDir := filepath.Join(t.TempDir(), "data")

	svc, err := bootstrap.NewEntitlementsService(config.Config{Storage: config.StorageSQLite, DataDir: dataDir, PlansFile: path})
	if err != nil {
		t.Fatalf("NewEntitlements: %v", err)
	}
	if svc == nil {
		t.Fatal("NewEntitlements returned no service")
	}
	got, err := svc.CurrentEntitlements(context.Background())
	if err != nil {
		t.Fatalf("CurrentEntitlements: %v", err)
	}

	export, ok := got.Features["export"]
	if got.Plan != "local" || !ok || export.IsOn || export.WhenOff != domain.ShowLocked || len(export.AvailableOn) != 0 {
		t.Errorf("CurrentEntitlements = %+v, want the local plan with export locked and no hint", got)
	}
	if _, err := os.Stat(dataDir); !os.IsNotExist(err) {
		t.Errorf("data folder stat error = %v, want it not created", err)
	}
}

// sample:start

// The shipped plans file has the sample's item export: on for the local and
// pro plans, and shown locked on free.
func TestNewEntitlements_ShippedPlansHaveItemExport(t *testing.T) {
	defer func(plan string) { bootstrap.LocalPlan = plan }(bootstrap.LocalPlan)
	for _, tc := range []struct {
		plan string
		isOn bool
	}{{"local", true}, {"pro", true}, {"free", false}} {
		t.Run(tc.plan, func(t *testing.T) {
			bootstrap.LocalPlan = tc.plan
			svc, err := bootstrap.NewEntitlementsService(config.Config{Storage: config.StorageMemory})
			if err != nil {
				t.Fatalf("NewEntitlements: %v", err)
			}
			got, err := svc.CurrentEntitlements(context.Background())
			if err != nil {
				t.Fatalf("CurrentEntitlements: %v", err)
			}

			export, ok := got.Features["item-export"]
			if !ok || export.Name != "Item export" || export.IsOn != tc.isOn || export.WhenOff != domain.ShowLocked {
				t.Errorf("item-export = %+v (found %t), want Item export, on %t, shown locked when off", export, ok, tc.isOn)
			}
		})
	}
}

// sample:end

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

// SQLite in a folder that looks synced gets a warning naming the service.
// Plain text there is supported (one device at a time), so it gets none, as
// does SQLite with the warning turned off, an ordinary folder, or a server.
func TestApp_WarnsWhenSQLiteDataLooksSynced(t *testing.T) {
	const dropbox = "Your data folder looks synced by Dropbox. A sync service can damage SQLite data: " +
		"move the data folder out of it, or keep your data as plain text."
	cases := []struct {
		name string
		cfg  func(synced string) config.Config
		want []string
	}{
		{"SQLite in a synced folder", func(synced string) config.Config {
			return config.Config{Storage: config.StorageSQLite, DataDir: synced}
		}, []string{dropbox}},
		{"plain text in a synced folder", func(synced string) config.Config {
			return config.Config{Storage: config.StorageJSONFile, DataDir: synced}
		}, nil},
		{"SQLite in a synced folder, warning turned off", func(synced string) config.Config {
			return config.Config{Storage: config.StorageSQLite, DataDir: synced, ShouldIgnoreSyncedFolder: true}
		}, nil},
		{"SQLite in an ordinary folder", func(string) config.Config {
			return config.Config{Storage: config.StorageSQLite, DataDir: t.TempDir()}
		}, nil},
		{"a server", func(string) config.Config {
			return config.Config{ServerURL: "http://localhost:1"}
		}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, ".dropbox"), nil, 0o600); err != nil {
				t.Fatal(err)
			}
			a, err := bootstrap.New(tc.cfg(filepath.Join(root, "data")))
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			t.Cleanup(func() { _ = a.Close() })

			if !slices.Equal(a.Warnings, tc.want) {
				t.Errorf("Warnings = %q, want %q", a.Warnings, tc.want)
			}
		})
	}
}
