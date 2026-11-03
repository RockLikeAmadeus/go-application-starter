package acceptance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

// registerSyncedFolderSteps binds the steps of features/synced-folder.feature.
func (w *world) registerSyncedFolderSteps(sc *godog.ScenarioContext) {
	sc.Step(`^the desktop app keeps its data in SQLite in a folder synced by (\w+)$`, w.theDesktopAppKeepsItsDataInSQLiteInASyncedFolder)
	sc.Step(`^the desktop app keeps its data as plain text in a folder synced by (\w+)$`, w.theDesktopAppKeepsItsDataAsPlainTextInASyncedFolder)
	sc.Step(`^I turn off the synced folder warning$`, w.iTurnOffTheSyncedFolderWarning)
	sc.Step(`^the server keeps its data in SQLite in a folder synced by (\w+)$`, w.theServerKeepsItsDataInSQLiteInASyncedFolder)
	sc.Step(`^the desktop app warns that its data folder looks synced by "([^"]*)"$`, w.theDesktopAppWarnsItsFolderLooksSyncedBy)
	sc.Step(`^the server warns that its data folder looks synced by "([^"]*)"$`, w.theServerWarnsItsFolderLooksSyncedBy)
	sc.Step(`^the desktop app gives no warning$`, w.theDesktopAppGivesNoWarning)
	sc.Step(`^the command line warns that its data folder looks synced by "([^"]*)"$`, w.theCommandLineWarnsItsFolderLooksSyncedBy)
	sc.Step(`^the command line gives no warning$`, w.theCommandLineGivesNoWarning)
}

// syncedDataFolder makes a folder under root that looks synced by service,
// as that service leaves it on a user's computer, and returns a data folder
// inside it. Which paths and files mean a synced folder is unit-tested; this
// is one of each kind: a marker file (Dropbox), a marker folder (Syncthing)
// and a folder's name (OneDrive).
func syncedDataFolder(root, service string) (string, error) {
	var synced string
	switch service {
	case "Dropbox":
		synced = filepath.Join(root, "Dropbox")
		if err := os.MkdirAll(synced, 0o755); err != nil {
			return "", err
		}
		if err := os.WriteFile(filepath.Join(synced, ".dropbox"), nil, 0o644); err != nil {
			return "", err
		}
	case "Syncthing":
		synced = filepath.Join(root, "Sync")
		if err := os.MkdirAll(filepath.Join(synced, ".stfolder"), 0o755); err != nil {
			return "", err
		}
	case "OneDrive":
		synced = filepath.Join(root, "OneDrive")
	default:
		return "", fmt.Errorf("no synced folder for %q in the tests", service)
	}
	dataDir := filepath.Join(synced, "starter")
	return dataDir, os.MkdirAll(dataDir, 0o755)
}

// theDesktopAppKeepsItsDataInSQLiteInASyncedFolder chooses SQLite and a data
// folder that service syncs in the settings file, and opens the desktop app
// again, so it reads them.
func (w *world) theDesktopAppKeepsItsDataInSQLiteInASyncedFolder(service string) error {
	if w.local == nil {
		return errNoLocalComputer
	}
	if err := w.local.addSetting("storage", string(sqliteStorage)); err != nil {
		return err
	}
	return w.theDesktopAppKeepsItsDataAsPlainTextInASyncedFolder(service)
}

// theDesktopAppKeepsItsDataAsPlainTextInASyncedFolder chooses a data folder
// that service syncs in the settings file, leaving the storage as it is
// (plain text unless chosen), and opens the desktop app again.
func (w *world) theDesktopAppKeepsItsDataAsPlainTextInASyncedFolder(service string) error {
	if w.local == nil || w.reopen == nil {
		return errNoLocalComputer
	}
	dataDir, err := syncedDataFolder(w.local.dir, service)
	if err != nil {
		return err
	}
	if err := w.local.addSetting("data_dir", dataDir); err != nil {
		return err
	}
	return w.reopen()
}

func (w *world) iTurnOffTheSyncedFolderWarning() error {
	if w.local == nil {
		return errNoLocalComputer
	}
	return w.local.addSetting("ignore_synced_folder", "true")
}

// theServerKeepsItsDataInSQLiteInASyncedFolder starts the server again on
// SQLite, in a data folder that service syncs.
func (w *world) theServerKeepsItsDataInSQLiteInASyncedFolder(service string) error {
	if w.server == nil {
		return errNoServer
	}
	dataDir, err := syncedDataFolder(w.server.dataDir, service)
	if err != nil {
		return err
	}
	w.server.dataDir = dataDir
	return w.server.restart(sqliteStorage)
}

func (w *world) theDesktopAppWarnsItsFolderLooksSyncedBy(service string) error {
	return warnsSyncedBy(w.apiURL, service)
}

func (w *world) theServerWarnsItsFolderLooksSyncedBy(service string) error {
	if w.server == nil {
		return errNoServer
	}
	return warnsSyncedBy(w.server.url, service)
}

func (w *world) theDesktopAppGivesNoWarning() error {
	got, err := warnings(w.apiURL)
	if err != nil {
		return err
	}
	if len(got) != 0 {
		return fmt.Errorf("the app warns %q, want no warning", got)
	}
	return nil
}

func (w *world) theCommandLineWarnsItsFolderLooksSyncedBy(service string) error {
	if want := "looks synced by " + service; !strings.Contains(w.cliErrOut, want) {
		return fmt.Errorf("the command line warned:\n%s\nwant it to say %q", w.cliErrOut, want)
	}
	return nil
}

func (w *world) theCommandLineGivesNoWarning() error {
	if w.lastErr != nil {
		return fmt.Errorf("the command line failed: %w", w.lastErr)
	}
	if w.cliErrOut != "" {
		return fmt.Errorf("the command line warned:\n%s\nwant no warning", w.cliErrOut)
	}
	return nil
}

// warnsSyncedBy checks that the app at url warns, as its web app shows, that
// its data folder looks synced by service.
func warnsSyncedBy(url, service string) error {
	got, err := warnings(url)
	if err != nil {
		return err
	}
	want := "looks synced by " + service
	for _, message := range got {
		if strings.Contains(message, want) {
			return nil
		}
	}
	return fmt.Errorf("the app warns %q, want a warning saying %q", got, want)
}

// warnings asks the app at url for its warnings, as its web app does.
func warnings(url string) ([]string, error) {
	resp, err := http.Get(url + "/api/v1/warnings")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("asking for the app's warnings: %s", resp.Status)
	}
	var body struct {
		Warnings []struct {
			Message string `json:"message"`
		} `json:"warnings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("reading the app's warnings: %w", err)
	}
	messages := make([]string, len(body.Warnings))
	for i, warning := range body.Warnings {
		messages[i] = warning.Message
	}
	return messages, nil
}
