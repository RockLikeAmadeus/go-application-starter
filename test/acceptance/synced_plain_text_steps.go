package acceptance

import (
	"os"
	"path/filepath"
	"time"

	"github.com/cucumber/godog"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/syncedfolder"
)

// registerSyncedPlainTextSteps binds the steps of
// features/synced-plain-text.feature.
func (w *world) registerSyncedPlainTextSteps(sc *godog.ScenarioContext) {
	sc.Step(`^a sync service left the conflict copy "([^"]*)" in the data folder$`, w.aSyncServiceLeftTheConflictCopy)
	sc.Step(`^I delete the conflict copy "([^"]*)"$`, w.iDeleteTheConflictCopy)
	sc.Step(`^the desktop app warns about the conflict copy "([^"]*)"$`, w.theDesktopAppWarnsAboutTheConflictCopy)
	sc.Step(`^the command line warns about the conflict copy "([^"]*)"$`, w.theCommandLineWarnsAboutTheConflictCopy)
	sc.Step(`^the device "([^"]*)" marked the data folder as in use (\d+) (minute|hour)s? ago$`, w.theDeviceMarkedTheDataFolder)
	sc.Step(`^the desktop app warns that the data folder looks in use on "([^"]*)"$`, w.theDesktopAppWarnsTheFolderLooksInUseOn)
	sc.Step(`^the command line warns that the data folder looks in use on "([^"]*)"$`, w.theCommandLineWarnsTheFolderLooksInUseOn)
}

// conflictCopyFile is a conflict copy's path, relative to the data folder. It
// is in a table of its own, so the template's scenarios need no sample.
func conflictCopyFile(name string) string {
	return filepath.Join("notes", name+".json")
}

// aSyncServiceLeftTheConflictCopy writes the file as a sync service would
// bring it: a copy keeps the original's contents, so its key isn't its name.
func (w *world) aSyncServiceLeftTheConflictCopy(name string) error {
	path, err := w.conflictCopyPath(name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(`{"version": 2, "key": "note", "record": {}}`), 0o644)
}

func (w *world) iDeleteTheConflictCopy(name string) error {
	path, err := w.conflictCopyPath(name)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

func (w *world) theDesktopAppWarnsAboutTheConflictCopy(name string) error {
	return warnsSaying(w.apiURL, conflictCopyFile(name))
}

func (w *world) theCommandLineWarnsAboutTheConflictCopy(name string) error {
	return w.commandLineWarnsSaying(conflictCopyFile(name))
}

// theDeviceMarkedTheDataFolder leaves the mark a sync service would bring from
// the device, dated as that device wrote it.
func (w *world) theDeviceMarkedTheDataFolder(device string, age int, unit string) error {
	dataDir, err := w.desktopAppsDataFolder()
	if err != nil {
		return err
	}
	ageUnit := time.Minute
	if unit == "hour" {
		ageUnit = time.Hour
	}
	return syncedfolder.WriteMark(dataDir, device, time.Now().Add(-time.Duration(age)*ageUnit))
}

// looksInUseOn is what a warning about device says.
func looksInUseOn(device string) string {
	return "appears to be in use on " + device
}

func (w *world) theDesktopAppWarnsTheFolderLooksInUseOn(device string) error {
	return warnsSaying(w.apiURL, looksInUseOn(device))
}

func (w *world) theCommandLineWarnsTheFolderLooksInUseOn(device string) error {
	return w.commandLineWarnsSaying(looksInUseOn(device))
}

// conflictCopyPath is where a conflict copy is in the desktop app's data
// folder.
func (w *world) conflictCopyPath(name string) (string, error) {
	dataDir, err := w.desktopAppsDataFolder()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, conflictCopyFile(name)), nil
}
