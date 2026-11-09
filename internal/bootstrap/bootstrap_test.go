package bootstrap_test

import (
	"context"
	"errors"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/jsonfile"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/syncedfolder"
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

func TestNewEmbedded_AnUnknownStorageErrorReadsAsASentence(t *testing.T) {
	_, err := bootstrap.NewEmbedded(config.Config{Storage: "disk"})

	want := `The storage setting "disk" is not recognized. Please use "memory", "jsonfile", or "sqlite".`
	if err == nil || err.Error() != want {
		t.Errorf("err = %v, want %q", err, want)
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

// The local edition's plan is set when the app is built, so asking for billing
// there is refused, and it opens no storage: it creates no data folder and
// isn't stopped by a locked one.
func TestLockedFolderError_OffersTheAlternativeForALockedFolderOnly(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{Storage: config.StorageJSONFile, DataDir: dir}
	a, err := bootstrap.NewEmbedded(cfg)
	if err != nil {
		t.Fatalf("NewEmbedded: %v", err)
	}
	t.Cleanup(func() { _ = a.Close() })

	_, err = bootstrap.NewEmbedded(cfg)
	got := bootstrap.LockedFolderError(fmt.Errorf("open: %w", err), "try this")

	want := "Your data folder is in use by another program on this computer (" + dir + "). Please close the other program and try again, or try this."
	if got == nil || got.Error() != want {
		t.Errorf("error = %v, want %q", got, want)
	}
	if !errors.Is(got, jsonfile.ErrLocked) {
		t.Errorf("error = %v, want it to match ErrLocked", got)
	}
	if got := bootstrap.LockedFolderError(errors.New("another error"), "try this"); got != nil {
		t.Errorf("another error gave %v, want nil", got)
	}
}

func TestNewBilling_RefusesInTheLocalEditionWithoutOpeningStorage(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")

	_, err := bootstrap.NewBilling(config.Config{Storage: config.StorageSQLite, DataDir: dataDir})

	if !errors.Is(err, bootstrap.ErrLocalPlanIsFixed) {
		t.Errorf("err = %v, want ErrLocalPlanIsFixed", err)
	}
	if _, err := os.Stat(dataDir); !os.IsNotExist(err) {
		t.Errorf("data folder stat error = %v, want it not created", err)
	}
}

// The refusal is shown to whoever runs the command, as it stands.
func TestNewBilling_LocalEditionRefusalWording(t *testing.T) {
	_, err := bootstrap.NewBilling(config.Config{Storage: config.StorageMemory})

	want := "The plan for the local edition is set when the application is built and cannot be changed."
	if err == nil || err.Error() != want {
		t.Errorf("err = %v, want %q", err, want)
	}
}

// The admin command works beside the hosted edition's storage and no endpoint
// changes plans, so asking for billing over a server is refused.
func TestNewBilling_RefusesOverAServer(t *testing.T) {
	_, err := bootstrap.NewBilling(config.Config{ServerURL: "http://localhost:1"})

	if !errors.Is(err, bootstrap.ErrNoPlanChangesOverAServer) {
		t.Errorf("err = %v, want ErrNoPlanChangesOverAServer", err)
	}
}

// The refusal is shown to whoever runs the command, as it stands.
func TestNewBilling_OverAServerRefusalWording(t *testing.T) {
	_, err := bootstrap.NewBilling(config.Config{ServerURL: "http://localhost:1"})

	want := "A plan cannot be changed through a server connection."
	if err == nil || err.Error() != want {
		t.Errorf("err = %v, want %q", err, want)
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
	const dropbox = "Your data folder appears to be synced by Dropbox. A file sync service can damage SQLite data: " +
		"please move your data folder out of the synced location before continuing. Alternatively, you can " +
		"switch to plain-text storage, but your existing data will not be transferred automatically."
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
			a, err := bootstrap.New(tc.cfg(dropboxDataDir(t)))
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

// A conflict copy is itself the evidence of a clash, so plain text warns about
// it, naming each file, wherever the folder is and whatever the synced folder
// setting says.
func TestApp_WarnsAboutConflictCopiesOnPlainText(t *testing.T) {
	first := filepath.Join("notes", "1 (conflicted copy).json")
	second := filepath.Join("notes", "2 (conflicted copy).json")
	one := "Your data folder contains 1 conflicted copy, " + first + ", which appears to have been created by " +
		"a file sync service. A change may be missing from your data. Please compare it with the original file, " +
		"then either delete it or use it to replace the original before continuing."
	many := "Your data folder contains 2 conflicted copies that appear to have been created by a file sync service: " +
		first + ", " + second + ". Changes may be missing from your data. Please compare each one with its original " +
		"file, then either delete it or use it to replace the original before continuing."
	cases := []struct {
		name         string
		copies       []string
		ignoreSynced bool
		want         []string
	}{
		{name: "one copy", copies: []string{first}, want: []string{one}},
		{name: "several copies", copies: []string{first, second}, want: []string{many}},
		{name: "synced folder warning turned off", copies: []string{first}, ignoreSynced: true, want: []string{one}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.Mkdir(filepath.Join(dir, "notes"), 0o755); err != nil {
				t.Fatal(err)
			}
			for _, name := range tc.copies {
				record := []byte(`{"version": 2, "key": "1", "record": {}}`)
				if err := os.WriteFile(filepath.Join(dir, name), record, 0o600); err != nil {
					t.Fatal(err)
				}
			}

			a, err := bootstrap.New(config.Config{
				Storage:                  config.StorageJSONFile,
				DataDir:                  dir,
				ShouldIgnoreSyncedFolder: tc.ignoreSynced,
			})
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

// A mark says another device has the data open. Plain text warns, naming the
// device, and still opens: the mark is a best guess.
func TestApp_WarnsWhenAnotherDeviceIsUsingPlainText(t *testing.T) {
	got := warningsWithMarksFrom(t, "work-laptop")

	want := []string{"Your data folder appears to be in use on work-laptop. Running the application from two " +
		"devices at once while plain-text storage is enabled can result in lost changes: please close the " +
		"application on your other device before continuing."}
	if !slices.Equal(got, want) {
		t.Errorf("Warnings = %q, want %q", got, want)
	}
}

// Each device could be the one to close, so the warning names them all.
func TestApp_WarningNamesEveryOtherDevice(t *testing.T) {
	got := warningsWithMarksFrom(t, "work-laptop", "work-phone")

	want := []string{"Your data folder appears to be in use on work-laptop and work-phone. Running the application " +
		"from multiple devices at once while plain-text storage is enabled can result in lost changes: please close " +
		"the application on your other devices before continuing."}
	if !slices.Equal(got, want) {
		t.Errorf("Warnings = %q, want %q", got, want)
	}
}

// Three or more names are a list, so they take the Oxford comma.
func TestApp_WarningListsThreeOtherDevicesWithTheOxfordComma(t *testing.T) {
	got := warningsWithMarksFrom(t, "home-pc", "work-laptop", "work-phone")

	want := []string{"Your data folder appears to be in use on home-pc, work-laptop, and work-phone. Running the " +
		"application from multiple devices at once while plain-text storage is enabled can result in lost changes: " +
		"please close the application on your other devices before continuing."}
	if !slices.Equal(got, want) {
		t.Errorf("Warnings = %q, want %q", got, want)
	}
}

// This computer's own mark is left by the app itself, perhaps a moment ago by
// another copy of it, so it isn't a reason to warn.
func TestApp_ThisDevicesOwnMarkGivesNoWarning(t *testing.T) {
	device, err := os.Hostname()
	if err != nil {
		t.Skipf("this computer's name can't be read: %v", err)
	}

	if got := warningsWithMarksFrom(t, device); len(got) != 0 {
		t.Errorf("Warnings = %q, want none", got)
	}
}

// dropboxDataDir returns a data folder inside a folder that Dropbox appears to
// keep in sync.
func dropboxDataDir(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".dropbox"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(root, "data")
}

// warningsWithMarksFrom opens plain text in a folder with a recent mark from
// each device, and returns the app's warnings.
func warningsWithMarksFrom(t *testing.T, devices ...string) []string {
	t.Helper()
	dir := t.TempDir()
	for _, device := range devices {
		if err := syncedfolder.WriteMark(dir, device, time.Now()); err != nil {
			t.Fatal(err)
		}
	}

	a, err := bootstrap.New(config.Config{Storage: config.StorageJSONFile, DataDir: dir})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = a.Close() })
	return a.Warnings
}

// A plain-text folder that looks synced gets this device's mark on opening, so
// that other devices can warn about it.
func TestApp_MarksAPlainTextFolderThatLooksSynced(t *testing.T) {
	device, err := os.Hostname()
	if err != nil {
		t.Skipf("this computer's name can't be read: %v", err)
	}
	dir := dropboxDataDir(t)

	a, err := bootstrap.New(config.Config{Storage: config.StorageJSONFile, DataDir: dir})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = a.Close() })

	got := syncedfolder.OtherDevices(dir, "another-device", time.Now())
	if want := []string{device}; !slices.Equal(got, want) {
		t.Errorf("devices with a mark = %q, want %q", got, want)
	}
}

// Only a folder that looks synced is marked, so the default folder, servers
// and Android have no background writes.
func TestApp_LeavesNoMarkInAFolderThatDoesNotLookSynced(t *testing.T) {
	if _, err := os.Hostname(); err != nil {
		t.Skipf("this computer's name can't be read, so no mark is left anyway: %v", err)
	}
	dir := filepath.Join(t.TempDir(), "data")

	a, err := bootstrap.New(config.Config{Storage: config.StorageJSONFile, DataDir: dir})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = a.Close() })

	if got := syncedfolder.OtherDevices(dir, "another-device", time.Now()); len(got) != 0 {
		t.Errorf("devices with a mark = %q, want none", got)
	}
}

// ignore_synced_folder says the folder isn't to be treated as synced, so a
// folder that looks synced gets no mark either.
func TestApp_LeavesNoMarkWhenTheSyncedFolderIsIgnored(t *testing.T) {
	if _, err := os.Hostname(); err != nil {
		t.Skipf("this computer's name can't be read, so no mark is left anyway: %v", err)
	}
	dir := dropboxDataDir(t)

	a, err := bootstrap.New(config.Config{Storage: config.StorageJSONFile, DataDir: dir, ShouldIgnoreSyncedFolder: true})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = a.Close() })

	if got := syncedfolder.OtherDevices(dir, "another-device", time.Now()); len(got) != 0 {
		t.Errorf("devices with a mark = %q, want none", got)
	}
}

// A scan that fails can't say there are no conflict copies, so plain text says
// it couldn't look, and the app still opens: the scan guards the data rather
// than being needed to reach it.
func TestApp_WarnsWhenConflictCopiesCantBeCheckedAndKeepsWorking(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory permissions don't make a folder unreadable on windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root can read an unreadable directory")
	}
	dir := t.TempDir()
	table := filepath.Join(dir, "notes")
	if err := os.Mkdir(table, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(table, 0o755) }) // so t.TempDir can remove it

	a, err := bootstrap.New(config.Config{Storage: config.StorageJSONFile, DataDir: dir})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = a.Close() })

	if len(a.Warnings) != 1 ||
		!strings.HasPrefix(a.Warnings[0], "Your data folder could not be checked for conflicted copies, so a change may be missing from your data. The following error occurred: ") ||
		!strings.Contains(a.Warnings[0], "notes") {
		t.Errorf("Warnings = %q, want one saying the check failed and naming the notes folder", a.Warnings)
	}
	if err := a.Ready(t.Context()); err != nil {
		t.Errorf("Ready error = %v, want the app to keep working", err)
	}
}
