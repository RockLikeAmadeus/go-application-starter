package acceptance

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
)

// registerSettingsFileSteps binds the steps of features/settings-file.feature,
// which only the desktop driver runs: the settings file is on the desktop
// app's computer, in its settings folder.
func (w *world) registerSettingsFileSteps(sc *godog.ScenarioContext) {
	sc.Step(`^my settings file chooses the data folder "([^"]*)"$`, w.mySettingsFileChoosesTheDataFolder)
	sc.Step(`^my settings file has the setting "([^"]*)" set to "([^"]*)"$`, w.mySettingsFileHasTheSetting)
	sc.Step(`^I start with the data folder "([^"]*)" chosen in the environment$`, w.iStartWithTheDataFolderInTheEnvironment)
	sc.Step(`^the desktop app keeps its data in "([^"]*)"$`, w.theDesktopAppKeepsItsDataIn)
	sc.Step(`^it says the data folder was chosen by "([^"]*)"$`, w.itSaysTheDataFolderWasChosenBy)
	sc.Step(`^I am told the settings file is invalid, because of "([^"]*)"$`, w.iAmToldTheSettingsFileIsInvalid)
}

func (w *world) mySettingsFileChoosesTheDataFolder(name string) error {
	if w.local == nil {
		return errNoLocalComputer
	}
	return w.local.setSetting("data_dir", filepath.Join(w.local.dir, name))
}

func (w *world) mySettingsFileHasTheSetting(key, value string) error {
	if w.local == nil {
		return errNoLocalComputer
	}
	return w.local.setSetting(key, value)
}

func (w *world) iStartWithTheDataFolderInTheEnvironment(name string) error {
	if w.local == nil {
		return errNoLocalComputer
	}
	w.local.env["STARTER_DATA_DIR"] = filepath.Join(w.local.dir, name)
	return nil
}

func (w *world) theDesktopAppKeepsItsDataIn(name string) error {
	if w.local == nil {
		return errNoLocalComputer
	}
	// The open desktop app holds the lock in its data folder.
	lock := filepath.Join(w.local.dir, name, ".lock")
	if _, err := os.Stat(lock); err != nil {
		return fmt.Errorf("the desktop app isn't using %s: %w", filepath.Dir(lock), err)
	}
	return nil
}

func (w *world) itSaysTheDataFolderWasChosenBy(source string) error {
	return w.commandLineSaid(source)
}

func (w *world) iAmToldTheSettingsFileIsInvalid(key string) error {
	if w.lastErr == nil {
		return errors.New("the command line worked, want it to say the settings file is invalid")
	}
	if !strings.Contains(w.lastErr.Error(), key) {
		return fmt.Errorf("want the error to name %q, got: %w", key, w.lastErr)
	}
	return nil
}

// settingsFile is where the app looks for the settings file on this
// computer: config.toml in its folder in the settings folder.
func (c *localComputer) settingsFile() string {
	return filepath.Join(c.dir, "settings", "starter", "config.toml")
}

// setSetting sets key = "value" in the settings file, replacing the key's
// line if the file already has one, and creating the file if needed.
func (c *localComputer) setSetting(key, value string) error {
	path := c.settingsFile()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	text, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	line := fmt.Sprintf("%s = %s", key, strconv.Quote(value))
	keyLine := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key) + `\s*=.*$`)
	if keyLine.Match(text) {
		text = keyLine.ReplaceAllLiteral(text, []byte(line))
	} else {
		text = append(text, line+"\n"...)
	}
	return os.WriteFile(path, text, 0o644)
}
